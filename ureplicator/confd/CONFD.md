## ignore me
Generate confd
```
ls -al | awk '{print$9}' | while read line ; do echo "src =" '"'${line}'"' ; echo "dest =" '"/uReplicator/config/'${line}'.tmpl"' ; echo  ; done
{{ getenv "HOSTNAME" }}
```
