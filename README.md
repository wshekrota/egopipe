# egopipe

### What is Egopipe?

Conventional ETL minimalist pipeline written in Go for Elasticstack. Basically it extends the pipeline 
capability of logstash giving you a go environment to manipulate your doc. It has a minimalist approach 
to configuring, stop living with complexity

Format: ![egopipe logo](https://www.google.com/imgres?imgurl=https%3A%2F%2Fgolangforall.com%2Fassets%2Ftube2.svg&imgrefurl=https%3A%2F%2Fgolangforall.com%2Fen%2Fgopher-drawings.html&tbnid=OMB0gw9yicfL9M&vet=10CO0BEDMomwNqFwoTCODfsYGJte8CFQAAAAAdAAAAABAE..i&docid=Ges437lBH6SG0M&w=800&h=519&q=golang%20gopher%20graphics&client=ubuntu&ved=0CO0BEDMomwNqFwoTCODfsYGJte8CFQAAAAAdAAAAABAE)

![logo](https://github.com/wshekrota/egopipe/blob/main/logo.png)

### Why use Egopipe?

If you have ever used Elasticstack the wonderful analytics suite from Elastic you quickly realize that the 
ingestion engine logstash is its great short coming. On the good side it has so many options. On the bad side
with great complexity comes difficulty debugging the various settings. Many times we all need a simplistic 
straight forward approach. I have always been a fan of having multiple stripped down versions of a product
over one that basically has too much in it. If you dig deeper you find logstash is written using JRuby and depends 
on a JVM. I have never liked the potential performance issues related to such a configuration.
Nothing is more funny than the paradox in naming the logstash pipeline language 'painless'.

I wanted a way to get into go code at little or no cost. Because the filebeat code is written in go and provides a 
rather rich well designed base of functionality I wanted to depend on it. The unfortunate late discovered
fact was that filebeat has a dependency on logstash input. If you tried to write your own logstash you would have to 
understand the complexity of whatever encode/decode (lumberjack) happens on that socketed interface between
filebeat and logstash.

I chose to let logstash receive the input and launch my pipe. Little changes and the configuration seems 
intuitive. The only thing you change is a small json config file and add a cert to the directory
if running securely. 


## config file management

---

Config is minimalist JSON and currently few values defined. Will expand as needed.

Defaults are defined in code. Values in egopipe.conf will override those.

Uses maps to individually union the defaults with config file settings.


```

"target": protocol:hostname:port 	http:127.0.0.1:9200

"name": unique part of index name	egopipe

"user": user if secured				""

"password": password if secured     ""

```

---

## installation and or testing the pipe

---

Clone the repo down and make changes to the transform stage 'yourpipecode'. Compile 
it and start the setup process by copying the mentioned files to the logstash directory.
If you need simple testing deployment consider using containers. I have sample Dockerfiles
in my egopipe_containers repo on my github account. These will build a cache for you so 
you can start containers at will. Remember stopping and starting again destroys the 
existing index(s). (for now)

The egopipe install directory is /etc/logstash/conf.d. But we have to put the non 
pipe files out of logstash reach creating an ego sub directory there.
After compiling the binary which allows you to add go code in filter section, place
it here.(in ego) Along with it will reside egopipe.cfg which is json. Note the hostnames
here assume you followed the elastic documentation in naming the hosts etc. Or you can
use ip addresses, especially if configuring it insecure or no ssl. The pipeline.conf file 
is just a stripped down logstash pipe that inputs and calls pipe to launch my code.

```

Content example for egopipe.cfg

{ "Target": "https://node-1.elastic.test.com:9200","User":"elastic","Password":"pw" }

(you will probably want to use a user you have setup with kibana DSL)

```

Input will be coming from filebeats installed on your application. (which is gitlab in this 
case for me) Configure filebeats to send the selected logs to port 5000 of the host where you 
have logstash running.

If securing the cluster follow those elastic instructions below copying the cert to this same 
directory as ca.pem. If the egopipe.conf file exists here egopipe reads it and prioritizes 
its items over the default values. Last thing is the suggested logstash pipeline.conf 
file whose output stage invokes egopipe. Place it in the conf.d directory.

What belongs in /etc/logstash/conf.d/ego? Everything but pipeline.conf and .yml file.

* executable (conf.d/ego)
* configuration file (conf.d/ego)
* pipeline.conf (conf.d)
* ca cert (conf.d/ego) following the elastic cert instructions you get a ca.crt which you rename to ca.pem
* logstash.yml (/etc/logstash)

You can make this deploy simpler if using containers your Dockefile can copy these files
in place. This is the way I'm doing my testing. You can check these out on another repo
at my github account. I think you might also put links to your clone in the Dockerfile 
directory that is if you originally built the cache with them.

Since I am not testing Elasticsearch I only built a single node elastic. Obviously you could take 
that single node container and build it to a multinode k8s cluster. I'll ignore that 
for the purpose of this document. I built the containers elasticsearch first then logstash
so I could guess the IP addresses 172.17.0.2 being first then 3 etc. (by default behavior)

---


## dependencies:

---

    - filebeat as input (queuing) recommend spool, default is memory 4096. 
    - egopipe config file should be copied to /etc/logstash/conf.d/ego to claim elastic host target.
    - logstash jvm.options config in /etc/logstash should have upper memory set as in ..Xmx2g
    - you have configured the rest of your elastic cluster as in the appropriate document *

[great elastic document](https://www.elastic.co/blog/configuring-ssl-tls-and-https-to-secure-elasticsearch-kibana-beats-and-logstash#prepare-logstash) *

---


## Status

---

Functional - Beta version distribution

---

Filebeat sends the socket in some layered lumberjack protocol on top of TCP. This makes it difficult to receive data 
direct on a go socket. I am opting to use a mostly null logstash pipe to receive the filebeat input then to 
launch the go program using the 'pipe' plugin. In this way I do not have to deal with lumberjack.

Logstash Pipe

```

input {
   beats {
      port => 5000
   }
}
filter {
}
output {
   pipe {
      command => "/etc/logstash/conf.d/ego/egopipe"
   }
}

```

On the output side of logstash what we get is JSON for the docs traveling through the pipelines. 
Then entering the go code we can decode and process that doc. Decoded we have a map which is very 
easy to manipulate in golang. When done we encode the map back to json. Ultimately we  put the 
annotated doc in an elastic index by way of an API call. 


## Transform stage or The Nitty Gritty (or how do I write the filter section in go)

This is where you take action on all or selectively some of the docs passing through the pipeline.
Forget the idea of plugins. Now you will effect changes in a go like manner. You will find it easy 
to do some things in a manner very similar to the comparitive logstash plugin or in many cases 
invent some new process.

I highly advise that we carry forward some of the same practices that made our pipe more easily 
debugable. Most forward in my mind is tag. Pipes are a form of headless operation thus difficult
to debug. Setting eye catchers is something you can design into your code. In logstash I had a tag 
for every nook and cranny. If something went wrong with specific logic I could look for the tag 
associated with it and find its fields using kibana. This is an indespensible practice.
Not every thing you will do needs to be documented here obviously. This is meant to get you started.

Your doc is a folder of fields. A field like message is a string containing the entire content.
You then act on that map to change and/or add new fields. Then we write it to the elastic index.


### Create a field or tag
```
h["tag"] = "Data for that field value"
```

### Does a field exist?
```
if _, ok := h["fieldname"]; ok {
  code for  field exist
} else
```

### Delete a field
```
delete(h,"fieldname")
```

### Rename a field
```
h["newfieldname"] = h["oldfieldname"]
delete(h,"oldfieldname")
```

### JSON decode (no prefixing string)
```
json.Unmarshal([]byte(h["message"].(string)),&h)
```

### JSON field (which has prefix not JSON will be left alone)
```
idx := strings.IndexRune(h["message"].(string),'{')
if idx>0 { json.Unmarshal([]byte((h["message"].(string))[idx:]),&h) }
```

### Suggested patterns for coding
```
A map is the object flowing through your pipe. This is decoded json. When you write
code here it may be difficult to decode the runtume version because pipes are headless.
In difficult times it may be helpful to use log.Println() to give you clues in the log 
at "/var/log/logstash/egopipe". Also I would consider defining a group of constant like
flags ie all camelcase to set the value for field tags. Tags is an array so append 
will work.

Where json was decoded as above...
H is defined map[string]interface{} as it is output from unmarshal. 
Likely tags field is already defined coming into the pipeline at filebeat, so you may 
append to it like this .. Perhaps you want to check does tags exist? Always be sure 
you are not operating on unintended docs flowing through your pipe. If you do, the runtime 
will obviously panic.
If we were to use "log.file.path" to be sure log was "/var/log/gitlab/gitaly/current"
it would make life more safe. You could also check for the existance of "tags" field 
using a suggestion above.
But never fear it gets more complex with fields using the '.' being submapped.
ie log.file.path would be defined by map["log":map["file":map["path":string ....

    x := h["log"].(map[string]interface{})
    y := x["file"].(map[string]interface{})
    z := y["path"].(string)
    if z == "/var/log/gitlab/gitaly/current" {
		h["tags"] = append(h["tags"].([]string),"DecodedJsonToFields")
	}

 or if tags does not exist

h["tags"] = []string{"DecodedJsonToFields"}

Note: I have asserted that tags value is an array of string.
Then later when looking at the pipeline data created in Kibana you will know what happened
in that doc.
```

### Extras - functions for transform stage


```
### dotField (h map[string]interface{}, field string)

Process fieldnames that have '.'. Some are simply field:value.
Others are submapped as "log.file.path" being map["log":map["file":map["path"]]].

```


```
### addTags(doc *map[string]interface{}, tags []string)

Add array of tags to doc. This lets you annotate different sections of code in data so 
that you can go back and make the association with some bad data.

```

```
### oniguruma(new_field string, field string, h map[string]interface{}, regex string)

Create new field from regex capture..

```


```
### grok(h *map[string]interface{}, regex string)

Create fields from patterns defined by %{syntax:semantic} patterns

```


## An example decode application in egopipe using gitlab logs

---

By filebeat feeding ../gitlab/gitaly logs  to our logstash we can experiment. The gitaly logs are completely 
json and we can use the top Unmarshal line above to decode the message into fields.

![kibana_screen](https://github.com/wshekrota/egopipe/blob/main/jsondecode.png)

In this screen you can look at the unchanged message field and that will tell you what decoded. Then look
at some of the fields on the left and notice they have decoded realtime ie. they are in the doc. Now to
take this a step further we could reference these fields directly. I would only caution that your code be 
written so you take actions when the field in question exists only. The sky is the limit for your designs.
You need only know to some degree what your objective is and what can be found in each type of log you 
might ingest. This takes practice and  experimentation.

This assumes the logs have some design in mind, they are json. In some cases you have totally raw logs.
These require ripping apart to get fields. If you know user login is in a raw log you have to design a 
regex type operation that once it identifies this is the right type log pulls out that substring and creates
a new field for it. Then forever more those type of doc will have this extra user field.

Consider that  once you have all these new fields created, in this gitlab case want to identify some 
operation like repo create? I know  that this gitaly log contains this from previous experimentation.
So I created a search in kibana that would identify that one operation. This is actually a combination
of an exact match identified by the use of 'keyword' (or non analyzed field) and a full text search
"CreateRepository". Be as specific as necessary, I prefer exact matches they are much faster.

```
grpc.request.glProjectPath.keyword:"wshek/newone11" AND "CreateRepository"
```

Once we are ingesting a given log we can then go to Kibana and experiment with fields to figure out 
what magic gives us what we want in result. It might be that when we identify some specific condition
we make things easier by creating a field in that specific timed doc. It's all a matter of design.
Should we do our work in Kibana or annotate the doc realtime in the pipe to make later operations easier?

---

## This is how output stage updates Elastic index

---

* POST /target/_doc/ 

Request body will carry the JSON.

Entire pipe is golang so transform or filter stage is familiar.

---

## Main errors on input or output stream

```

    Config
       good: use config
       bad: no file continue default
       bad: file decode error SystemExit
    JSON decode stream
       good: nil
       bad: SystemExit
    JSON encode stream
       good: nil
       bad: SystemExit
    POST call
       good:
       Response body 
       How to evaluate _shards or determine success.
       bad: SystemExit

```


## logstash to egopipe

![flowchart](https://github.com/wshekrota/egopipe/blob/main/egopipe.png)

---

**egopipe** will:

**stage1**

* read stdin
* decode that json

**stage2** (called w/ map)

* write user pipeline

**stage3** (called w/ ref to stage2 map)

* encode back to json
* POST API call (write to elastic index)

---


## Security considerations

```

Enabling TLS on an Elastic transaction is a two fold process. We must setup TLS over the http
connection. So we copy the ca cert to the /etc/logstash/conf.d/ego along with the egopipe executable
and conf file. The conf file MUST contain a target that specifies https not http and a user
and password to authenticate. Both TLS and authentication are a requirement.


```


## Testing egopipe

Integrated go testing tests included in repository. Tests config file and output stage so far.

Alternate test
```
Echo some JSON into the executable..

ex: echo 'some json' | ./egopipe

Normally the executable is invoked by stage3 logstash pipeline you have in place.

example doc: (json)

{"@timestamp":"2021-03-10T00:58:21.976Z","tags":["beats_input_codec_plain_applied"],"ecs":{"version":"1.4.0"},"agent":{"hostname":"gitlab","id":"3fc83c01-53ee-4a26-b33a-fe2ec2e4dab6","type":"filebeat","version":"7.6.2","ephemeral_id":"55da0cbe-abff-4ec2-9820-624ceff143c1"},"input":{"type":"log"},"log":{"offset":0,"file":{"path":"/var/log/walt/walt.log"}},"host":{"architecture":"x86_64","name":"gitlab","containerized":false,"os":{"kernel":"4.15.0-99-generic","name":"Ubuntu","version":"18.04.4 LTS (Bionic Beaver)","codename":"bionic","platform":"ubuntu","family":"debian"},"hostname":"gitlab","id":"f1823a3617ff4004baa550a0ae8408b6"},"message":"yessir","@version":"1"}
```


## Metrics

```

These metrics give you a realtime look at what might be happening with your data. If you want 
something more exotic you should use Kibana. Figures are post filter so additions/deletions will be seen.

Elapsed - rough estimate time after read/decode to write index success (end to end)
Bytes - total bytes transit
Docs - total doc count
Fields - map is docs grouped by number of fields per doc and count within ex: map[27:2,11:14] would be there are 2 docs w/27 fields etc. 

* this gives you a rough idea what is coming in and what is done in the filter.

Rest endpoint (todo maybe)
GET list - all docs
GET id= - individual doc

```


## TODO list

---

add target config value (complete 03/12)

add name config value (complete 03/12)

config changes now installed /etc/logstash/conf.d (complete 03/15)

metrics What kind? (complete 03/21)

metrics wants perhaps an end to end traversal time (complete 03/21)

change default location of log and name /var/log/logstash/egopipe/egopipe (complete 03/22)

security (add user/pw to config for elastic) (committed 03/24-26) testing

assess comparison logstash plugins to go methods (README)

test what happens in queue backup

add metrics as a rest API call... GET metrics?

evaluate debug function output

now that I have a working model refactor code into separate files organized within package (complete 03/28)

refactor pipe read to read all before process so map[string]interface{} becomes []map....

suggestion of persistence (since that would be outside egopipe code in elastic)

rethink index name defaults this may interfere with what is expected


---
