Commit Count

This simple script fetch git repo and do a count on the number of commits made by each user. 

Edit setting.yml to put in repo and contributors. 
Repo names must be one word.

Then execute
```
$ bin/commit-count
```

The script will clone all repo or update if directory already exists concurrently. After execution finishes, result file will be stored in work/result.csv. 
