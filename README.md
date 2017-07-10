# Description

**mmdu** - Mysql Manager of Databases and Users.  
This service will ensure, that only users and databases defined in config file will exist. Everything which is not in there will be dropped, everything which is missing - added.  
Default behavior is to keep data untouched but report about differences.  
Config file can be generated via puppet, chef, salt, ansible  or any other config management system.  

Inspired by puppet mysql module, but without known architectural problems.  

Effect on freshly-installed mysql-host with **mmdu** config from this repository will look like this:  
```
$ mmdu # Run to check  
DROP DATABASE test  
CREATE DATABASE oleg_test  
CREATE DATABASE qwerty  
DROP USER 'root'@'127.0.0.1'  
DROP USER 'root'@'::1'  
DROP USER ''@'aw-db01.oleg'  
DROP USER 'root'@'aw-db01.oleg'  
DROP USER ''@'localhost'  
DROP USER 'debian-sys-maint'@'localhost'  
DROP USER 'root'@'localhost'  
GRANT SELECT, UPDATE ON `oleg%`.* TO 'oleg'@'10.%' IDENTIFIED BY PASSWORD '*F41E614E894A46E0FB7317B1C8CB6CEA97415C7B'  
GRANT SELECT ON qwerty.* TO 'oleg'@'10.%' IDENTIFIED BY PASSWORD '*F41E614E894A46E0FB7317B1C8CB6CEA97415C7B'  
GRANT ALL PRIVILEGES ON *.* TO 'root'@'127.0.0.1' IDENTIFIED BY PASSWORD '*81F5E21E35407D884A6CD4A731AEBFB6AF209E1B' WITH GRANT OPTION  
GRANT ALL PRIVILEGES ON *.* TO 'root'@'localhost' IDENTIFIED BY PASSWORD '*81F5E21E35407D884A6CD4A731AEBFB6AF209E1B' WITH GRANT OPTION  
$ mmdu -e # Run to execute  
$ mmdu # Run to check  
$  
```  
All actions happen in single transaction. If something went wrong - changes will not be applied.  
This is also make possible to change root-user password and not being disconnected.  

Second run will give us:  
```
$ mmdu # Same will be with mmdu -e in this case  
Nothing to do  
```  

# Options
 - **-e** - execute/apply all changes
 - **-c** - custom location of config file (default is */etc/mmdu/mmdu.toml*)

# Configuration

You need to update settings and add your own data into configuration file */etc/mmdu/mmdu.toml*:  

- **[general]** -general configuration section
  - autoExecute - behaves like `-e` parameter. If true - executes statements. Good for staging servers to automatically execute things. Default `false`
  
- **[access]** - this section contains data to connect to mysql server  
  You can specify *username*, *password*, *initPass*, *host*, *port*. If you do not - it has default values (root without password to localhost 3306)  
  - username - users which has permissions to grant (WITH GRANT OPTION) privileges, drop and create databases and users  
  - password - password for this user  
  - initPass - if your mysql was managed before by something else and has password, you can provide initial data to login and make changes. Default is no password authentication  
  - host - host where mmdu should connect. Default is "localhost"  
  - port - port where mmdu should connect. Default is 3306  
  - socket - preferred. If you have this parameter set up, mmdu will try to connect via socket. This is prefered, because in 5.7 this is the only way to connect with root to freshly installed DB  
  
- **[[database]]** - Databases which need to be in mysql.
  - name - name of database to manage
  
- **[[user]]** - Users which need to be in mysql.
  You can specify *username*, *network*, *password*, *hashedPassword*, *database*, *table*, *privileges*, *grantOption*  
  - username - username
  - network - address, from which mysql allows to connect this user. e.g. *8.8.%*  
  - password - plain-text password. I would not recommend to have it. Better use *hashedPassword*. If specifyed - *hashedPassword* will be ignored  
  - hashedPassword - hashed password (sha1 encrypted). You can get it via ```mysql> select password('oleg')``` e.g. **\*F41E614E894A46E0FB7317B1C8CB6CEA97415C7B**
  - grantOption - true or false flag for users "WITH GRANT OPTION". Default *false*  
  
- **[[user.permissions]]**
  - database - database to which user has an access. **mmdu** will ensure this database exist **[[database]]**. e.g. *oleg%* or **\***
  - table - table to which user has an access. e.g. "*"
  - privileges - list of privileges for user. e.g. ["SELECT", "INSERT"]

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
