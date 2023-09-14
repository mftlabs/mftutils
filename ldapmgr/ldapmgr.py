import json
import csv
import ldap3
import argparse, sys
import random
import string
from ldap3 import Server, Connection, ALL, NTLM, ALL_ATTRIBUTES, ALL_OPERATIONAL_ATTRIBUTES, AUTO_BIND_NO_TLS, SUBTREE

class LdapManager(object):
    def __init__(self, config_file, data_file, output_file):
        self.config = json.load(open(config_file))
        self.data_file = data_file
        self.output_file = output_file
        self.data = csv.reader(open(self.data_file))
        self.data = list(self.data)
        self.data = self.data[1:]
        self.connection = self.config['connection']
        self.base_dn = self.config['base_dn']
        self.bind_password = self.config['bind_password']
        self.ldap_connection = None
        self.entries = None
        self.login_shell = self.config['login_shell']
        self.home_directory = self.config['home_directory']
        self.uid_number = self.config['uid_number']
        self.gid_number = self.config['gid_number']
        self.search_dn = self.config['search_dn']
        self.batch_count = 0
        self.user_dn = self.config['user_dn']
        self.object_class = self.config['object_class']

    def connect(self):
        self.ldap_connection = self.connect_ldap(self.connection, self.base_dn, self.bind_password)
        print('Connection status: ',self.ldap_connection)

    def search(self, search_filter='(&(objectClass=inetOrgPerson))', search_scope=SUBTREE, attributes=[ALL_ATTRIBUTES, ALL_OPERATIONAL_ATTRIBUTES]):
        self.ldap_connection.search(search_base=self.search_dn, search_filter=search_filter, search_scope=search_scope, attributes=attributes)
        return self.ldap_connection.entries

    def search_by_uid(self, uid):
        self.ldap_connection.search(search_base=self.search_dn, search_filter='(&(objectClass=inetOrgPerson)(uid={}))'.format(uid), search_scope=SUBTREE, attributes=[ALL_ATTRIBUTES, ALL_OPERATIONAL_ATTRIBUTES])
        return self.ldap_connection.entries

    def generate_random_password(self, length=10):
        return ''.join(random.choice(string.ascii_uppercase + string.digits) for _ in range(length))


    def add_ou(self, ou):
        modlist = {}
        modlist['ou']=ou
        self.ldap_connection.add('ou={},ou=users,dc=hub4edi,dc=dev'.format(ou), ['top','organizationalUnit'], modlist)
        print(self.ldap_connection.result)

    def add_entries(self):
        print("Adding new entries into ldap")
        users_added = []
        for row in self.data:
            print('Processing: ',row)
            if len(row) >= 3:
                if len(self.search_by_uid(row[0])) == 0:
                    print('Adding: ',row)
                    self.batch_count += 1
                    user_password = self.generate_random_password()
                    self.ldap_connection.add(self.user_dn.format(row[0], row[1]), self.object_class, {'uid':row[0], 'sn': row[0], 'cn': row[0], 'userPassword': user_password, 'mail': row[2],'loginShell': self.login_shell, 'homeDirectory': self.home_directory.format(row[0]), 'uidNumber': str(self.uid_number)+str(self.batch_count), 'gidNumber': str(self.gid_number)+str(self.batch_count)})
                    print('Result of add operation: ',self.ldap_connection.result)
                    self.ldap_connection.search(search_base=self.base_dn, search_filter='(uid={})'.format(row[0]), search_scope=SUBTREE, attributes=[ALL_ATTRIBUTES, ALL_OPERATIONAL_ATTRIBUTES])
                    print('Inserted ',self.ldap_connection.entries)
                    row.append(user_password)
                    row.append(self.home_directory.format(row[0]))
                    row.append(str(self.uid_number)+str(self.batch_count))
                    row.append(str(self.gid_number)+str(self.batch_count))
                    users_added.append(row)
                else:
                    print('User already exists: ',row)
            else:
                print('Invalid row: ',row)
        print('Users added: ',users_added)
        with open(self.output_file, 'w') as f:
            writer = csv.writer(f)
            writer.writerows(users_added)

    def bind(self):
        self.ldap_connection.bind()

    def unbind(self):
        self.ldap_connection.unbind()

    def connect_ldap(self,connection, base_dn, bind_password):
        server = Server(connection, get_info=ALL)
        conn = Connection(server, user=base_dn, password=bind_password, auto_bind=True)
        return conn

def main():
    # Initialize parser
    msg = "Ldap Manager, a tool to manage ldap entries"
    parser = argparse.ArgumentParser(description=msg)
    parser.add_argument("-c", "--config", help = "Config File",  default = "config.json")
    parser.add_argument("-d", "--data", help = "Data File")
    parser.add_argument("-o", "--output", help = "Output File", required=True)
    parser.add_argument("-a", "--add", help = "Add new entry",action=argparse.BooleanOptionalAction, default = False)
    parser.add_argument("-s", "--search", help = "Search for entries",action=argparse.BooleanOptionalAction, default = False)
    parser.add_argument("-u", "--ou", help = "Arganizatioin unit")
    args = parser.parse_args()
    print(args.config, args.data)
    app = LdapManager(args.config, args.data, args.output)
    app.connect()
    app.bind()
    if args.add:
        app.add_entries()
    if args.search:
        data = app.search()
        print(data)
    if args.ou:
        app.add_ou(args.ou)
    app.unbind()
    print(app.entries)





if __name__ == '__main__':
    main()

