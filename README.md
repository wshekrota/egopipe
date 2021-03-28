# egopipe

### What is Egopipe?

Conventional ETL minimalist pipeline written in Go for Elasticstack. Basically

it extends the pipeline capability of logstash giving you a go environment to

manipulate your doc. It has a minimalist approach to configuring, stop living

with complexity

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

egopipe and egopipe.conf now should be installed to '/etc/logstash/conf.d'.

```

"target": "https://192.168.1.43:9200" (default)

"name": "egopipe" (default)

"user": "fred" (default is empty or insecure)

"password": "bogus99" (same)

```

---

## installation

---

Egopipe install directory is /etc/logstash/conf.d.
After compiling the binary which allows you to add go code in filter section, place
it here. Along with it will reside egopipe.conf which is json. Note the hostnames
here assume you followed the elastic documentation in naming the hosts etc.

```

{ "Target": "https://node-1.elastic.test.com:9200","User":"elastic","Password":"pw" }

(you will probably want to use a user you have setup with kibana DSL)

```

If securing the cluster follow those rules below copying the cert to this same 
directory. If the file exists here egopipe reads it and prioritizes its items over the 
default values.
Last thing is  the suggested logstash pipeline.conf file whose output stage 
invokes egopipe.

```

.l executatble
.l configuration file
.l pipeline.conf (logstash)
.l ca cert

```

---


## dependencies:

---

    - filebeat as input (queuing) recommend spool, default is memory 4096. 
    - egopipe config file should be copied to /etc/logstash/conf.d to claim elastic host target.
    - logstash jvm.options config in /etc/logstash should have upper memory set as in ..Xmx2g
    - you have configured the rest of your elastic cluster as in the appropriate document *

[great elastic document](https://www.elastic.co/blog/configuring-ssl-tls-and-https-to-secure-elasticsearch-kibana-beats-and-logstash#prepare-logstash)

---

## Status

---

Functional - currently in testing

---

Filebeat sends socket in some layered lumberjack protocol on top of TCP. This makes it difficult to receive data 
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
      command => "/etc/logstash/conf.d/egopipe"
   }
}

```

On the output side of logstash what we get is JSON for the docs traveling through the pipelines. Then entering the go code we 
can decode and process that doc. Decoded we have a map which is very easy to manipulate in golang.
When done we encode the map back to json and ultimately we  put the annotated doc in an elastic index by way of an API call. 


## The Nitty Gritty (or how do I write the filter section in go)

```

The logstash pipe filter section used to just be Ruby like code. Now the filter section or stage2 

of egopipe may be written in go. Of course when done egopipe must be compiled then copied to 

'/etc/logstash/conf.d'. The doc object is the map 'h' so there will be many operations you can 

perform similar to their logstash equivalents that I will document below.


h["name"] = "this is a test"  // add a field

delete(h,"key")               // delete a field

json.Unmarshal([]byte(h[key].(string)),&h)     // decode a json value for a key

                              // this processes top level json prefixed by non json
idx := strings.IndexRune(h["message"].(string),'{')   // json convert of message
if idx>0 { json.Unmarshal([]byte((h["message"].(string))[idx:]),&h) }

```

## This is how output stage updates Elastic index

---

* POST /target/_doc/ *

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
* PUT API call (write to elastic index)

---


## Security considerations

```

Enabling TLS on an Elastic transaction is a two fold process. We must setup TLS over the http
connection. So we copy the ca cert to the /etc/loggstash/conf.d along with the egopipe executable
annd conf file. The conf file MUST contain a target that specifies https not http and a user
and password to authenticate. Both TLS and authentication are a requirement.


```


## Testing egopipe

```
Echo some JSON into the executable..

ex: echo 'some json' | ./egopipe

Normally the executable is invoked by stage3 logstash pipeline you have in place.

example doc: (json)

{"@timestamp":"2021-03-10T00:58:21.976Z","tags":["beats_input_codec_plain_applied"],"ecs":{"version":"1.4.0"},"agent":{"hostname":"gitlab","id":"3fc83c01-53ee-4a26-b33a-fe2ec2e4dab6","type":"filebeat","version":"7.6.2","ephemeral_id":"55da0cbe-abff-4ec2-9820-624ceff143c1"},"input":{"type":"log"},"log":{"offset":0,"file":{"path":"/var/log/walt/walt.log"}},"host":{"architecture":"x86_64","name":"gitlab","containerized":false,"os":{"kernel":"4.15.0-99-generic","name":"Ubuntu","version":"18.04.4 LTS (Bionic Beaver)","codename":"bionic","platform":"ubuntu","family":"debian"},"hostname":"gitlab","id":"f1823a3617ff4004baa550a0ae8408b6"},"message":"yessir","@version":"1"}
```


## Metrics

```

These metrics give you a realtime look at what might be happening with your data. If you want something more exotic you should use Kibana. Figures are post filter so additions/deletions will be seen.

Elapsed - rough estimate time after read/decode to write index success (end to end)
Bytes - total bytes transit
Docs - total doc count
Fields - map is docs grouped by number of fields per doc and count within, this gives you a rough idea what is coming in and 
what is done in the filter.
docFields - number fields this doc

Rest endpoint (maybe)
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

security (add user/pw to config for elastic) (committed 03/24i-26) testing

assess comparison logstash plugins to go methods (README)

test what happens in queue backup

add metrics as a rest API call... GET metrics?

evaluate debug function output

now that I have a working model refactor code into separate files organized within package
---
