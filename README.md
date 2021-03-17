
GoLang Set Persistence Utility
==============================


WARNING: THIS ISN"T ATOMIC OR 100%

*Why a "Set Persistent" utility?*

I have some AWS go things I run to e.g. reencrypt all the objects in a
bucket or perhaps just re-copy them all.  My credentials expire in an
hour, but sometimes it takes longer than that to run. For re-encryption, the
state is in the S3 bucket, but for re-copying it's not as cut and dry
(maybe the timestamps but time is tricky) so this lets one quickly
make note of a bucket/object combo having been done and then once a
minute it flushes the state out to disk.  Out of an hour if I dup 1
minute of work, that's fine.

I just gathers a bunch of key names and writes them to a file as lines in a 
dedicated go routine.  It was using BoltDB. Initially I had BoltDB run 
synchronously with each copy but the process went from using all 
cores to just 1 core, mostly due to lock waiting for the DB transaction 
I think.  Even just writing out the data once per minute was a bit slow. 
Currently it doesn't seem to be a long poll in the re-copying process.

*Is it good to use?*

I'm using it.  

*What is it? *

Call "InSet" on a string and get a bool. 

Call "Set" with a string and eventually that answer will be
persisted to a flat file.  

*Who owns this code?*

Chris Lane

*Adivce for starting out*

If you integrate, please let me or them know of your experience and
any suggestions for improvement.

The current API can best be seen in the _test files probably.  

*Requirements*

BoltDB.  
