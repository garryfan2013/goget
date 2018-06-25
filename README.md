# goget

"goget" is a golang project that could be used for file downloading management

## Features:

1. Support http/ftp protocol
2. Multi-job management(status/progress query, stop/start control)
3. Multi-task(like thread,but actually goroutine) downloading for single job
4. Job resuming from break-point

## Howto

### build

<pre>
go get github.com/garryfan2013/goget
cd $GOPATH/src/garryfan2013/goget
go build
cd ./cli
go build
</pre>

### Server damon

<pre>
./goget
</pre>

### CLI

list all jobs
<pre>
./cli -L
</pre>

list single job
<pre>
./cli -l JOB_ID
</pre>

Add a job, this command succeeds will output the JOB_ID
<pre>
./cli -a URL -c TASK_COUNT -o SAVE_PATH -u USERNAME -p PASSWD
</pre>

Start a job
<pre>
./cli -s JOB_ID
</pre>

Stop a job
<pre>
./cli -S JOB_OD
</pre>

Delete a job
<pre>
./cli -d JOB_ID
</pre>

## Todo

1. Provide web interface for convinient management
2. Dynamically adjust task count according to statistics
3. Summary some golang specified knowledge and skill

## Special thanks

1. [RPC package - grpc](https://github.com/grpc/grpc)
2. [Ftp client package](https://github.com/jlaffaye/ftp)
3. [Simple kv store - boltDB](https://github.com/boltdb/bolt)
4. [Uuuid package - go.uuid](https://github.com/satori/go.uuid)