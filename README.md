****BeanBacker****

Utility to create encrypted backups of environment variables provided in AWS ElasticBeanstalk environments

---

**Usage**

```
go install github.com/mudasirmirza/beanbacker
```

```
# fetching all environment variables of all environments
beanbacker -destFile=./data.json -passPhrase=passwordtoencryptwith

# decrypting the fetched file
beanbacker -destFile=./data.json -passPhrase=passwordtoencryptwith -decryptData=true
```
