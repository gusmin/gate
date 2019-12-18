# AUTO INSTALLER SNORT
# OS
For the installation of snort, we are going to use Debian 9.x or younger versions.
# Preview:
# What is snort ?
## Snort is a free open source network intrusion detection system (IDS) and intrusion prevention system (IPS). He is there for warn you if someone is trying to stole your precious ping fluffy unicorns pictures for his malicious own usage. We can resume Snort as an Holy pig guard angel here for detect intrusion in your ssh kingdom
# Why do you need snort ?
## We use Snort in SecureGate for help you to secure your communication between the gate and your other machine... but it could be difficult for a newbie to aprehend a new technology without a little bit of help so here is a help install script and some tips who could be usefull to the hero you are.
# Install Script.
## The script 'installSnort.sh' will install snort and all dependencies than you need.
## usage:

```shell
  $ sudo installSnort.sh
```
## For check if everything is alright you can check your installation with the command:

```shell
  $ sudo snort -V
    ,,_     -*> Snort! <*-
  o"  )~   Version 2.9.15 GRE (Build 7) 
   ''''    By Martin Roesch & The Snort Team: http://www.snort.org/contact#team
           Copyright (C) 2014-2019 Cisco and/or its affiliates. All rights reserved.
           Copyright (C) 1998-2013 Sourcefire, Inc., et al.
           Using libpcap version 1.8.1
           Using PCRE version: 8.39 2016-06-14
           Using ZLIB version: 1.2.11

```

```shell
  $ cat /etc/snort/snort.conf
```
## By default Snort will track paquets on every ip address but you can configure it.
## For ask to your favorite snort pig to make his watch turn on a particular ip address you have to get your ip:
### dependent of your version of Debian you can use 'ifconfig' or 'ip addr' command;
```shell
$ ip addr
```
### Take your inet value (by example: 172.42.666.0/24 :))
### Now you have to edit the HOME_NET value in the file /etc/snort/snort.conf. You can replace the line
```shell
ipvar HOME_NET any
```
### by 
```shell
ipvar HOME_NET 172.42.666.0/24
```
### now for launch snort use the following command:
```shell
sudo snort -c /etc/snort/snort.conf -A console -l /var/log/snort -K Ascii
```

## The fantastic part begin now !!!
# With snort you can add your OWN RULES for more security
## But what is a rules ?
### A rules is a command line based on an sender adress and an receiver adress. It will analyse an certain kind of paquets (TCP, FTP....)
### There is a directory '/etc/snort/rules' where you will find every rules files. Every one of them contain a different category of rules.
### In the '/etc/snort/rules/local.rules' you can add your own rules with your personal alert message like those ones:

```shell
alert tcp any any -> $HOME_NET 21 (msg:"FTP connection attempt"; sid:1000001; rev:1;)
alert icmp any any -> $HOME_NET any (msg:"ICMP connection attempt"; sid:1000002; rev:1;)
alert tcp any any -> $HOME_NET 80 (msg:"Web connection attempt"; sid:1000003; rev:1;)
alert tcp any any -> $HOME_NET 22 (msg:"SSH connection attempt"; sid:1000004; rev:1;) 
```

### Those different rules will:
#### warn you about a FTP, ICMP, Web and SSH connexion.

