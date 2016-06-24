# Description

This is a service to manage mysql users and databases.  
Inspired by puppet mysql module, but without known architectural problems.



# Configuration

There is a config file which must be located under */etc/mmdu/mmdu.toml*  
But you can redefine it with option *-c*  
You need to update settings and add your own data into int:  

- **[access]** - this section contains data to connect to mysql server.  
  You can specify *username*, *password*, *initPass*, *host*, *port*. If you do not - it has default values (root without password to localhost 3306).  
  - username - users which has permissions to grant (WITH GRANT OPTION) privileges, drop and create databases and users  
  - password - password for this user  
  - initPass - we need it because by default mysql does not have root password or has 1 time pass (5.7). Default is no password authentication  
  
- **[[database]]** - Databases which need to be in mysql.
  - name - name of database to manage
  I strongly recommend you to keep **information_schema**, **performance_schema**, **mysql** in file. Without them mysql server will not work. 
  
- **[[user]]** - Users which need to be in mysql.
  You can specify *username*, *network*, *password*, *hashedPassword*, *database*, *table*, *privileges*, *grantOption*  
  - username - username
  - network - address, from which mysql allows to connect this user. e.g. *8.8.%*  
  - password - plain-text password. I would not recommend to have it. Better use *hashedPassword*. If specifyed - *hashedPassword* will be ignored  
  - hashedPassword - hashed password (sha1 encrypted). You can get it via ```mysql> select password('password')``` e.g. *\*F41E614E894A46E0FB7317B1C8CB6CEA97415C7B*  
  - database - database to which user has an access. **mmdu** will ensure this database exist **[[database]]**. e.g. *oleg%* or *\**
  - table - table to which user has an access. e.g. "*"
  - privileges - list of privileges for user. e.g. ["SELECT", "INSERT"]
  - grantOption - true or false flag for users "WITH GRANT OPTION". Default *false*

# Installation

- Install go https://golang.org/doc/install
- Make a proper structure of directories: ```mkdir -p /opt/go/src /opt/go/bin /opt/go/pkg```
- Setup g GOPATH variable: ```export GOPATH=/opt/go```
- Clone this project to src: ```go get github.com/leoleovich/mmdu```
- Fetch dependencies: ```cd /opt/go/github.com/leoleovich/mmdu && go get ./...```
- Compile project: ```go install github.com/leoleovich/mmdu```
- Copy config file: ```mkdir /etc/mmdu && cp /opt/go/src/github.com/leoleovich/mmdu/mmdu.toml /etc/mmdu/```
- Adjust your settings
- Run it ```/opt/go/bin/mmdu```
