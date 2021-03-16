9# egopipe
A minimalist solution for logstash complexity in Elastic

# Conventional ETL minimalist pipeline written in Go for Elasticstack

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


## config file management

---

Config is minimalist and currently few values defined. Will expand as needed.

Defaults are defined in code. Values in egopipe.conf will override those.

egopipe and egopipe.conf now should be installed to '/etc/logstash/conf.d'.

```

"target": "http://127.0.0.1:9200" (default)

"name": "egopipe" (default)

```

---


## dependencies:

---

    - filebeat as input (queuing) recommend spool, default is memory 4096. 
    - egopipe config file should be copied to /etc/logstash/conf.d to claim elastic host target.
    - logstash jvm.options config in /etc/logstash should have upper memory set as in ..Xmx2g

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

---

## The Nitty Gritty (or how do I write the filter section in go)

```

The logstash pipe filter section used to just be Ruby like code. Now the filter section or stage2 of egopipe may be written in go.

Of course when done egopipe must be compiled then copied to '/etc/logstash/conf.d'. The doc object is the map 'h' so there will be many

operations you can purform similar to their logstash equivallents that I will document below.

h["name"] = "this is a test"  // add a field

delete(h,"key")               // delete a field

json.Unmarshal([]byte(h[key].(string)),&h)     // decode a json value for a key

```

### POST /target/_doc/

---

Request body will carry the JSON.

Entire pipe is golang so transform or filter stage is familiar.

```

Errors:

    JSON decode
       good: nil
       bad: numeric error
    JSON encode
       good: nil
       bad: numeric error
    POST call
       Response body 
       How to evaluate _shards or determine success.

```


## logstash to egopipe

![flowchart](https://github.com/wshekrota/egopipe/egopipe.png)
---

input(socket) -> filter(null) -> output(pipe) | **input(stdin) -> filter(your go code) -> output(write index)**

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


## Testing egopipe

```
Echo some JSON into the executable..

ex: echo 'some json' | ./egopipe

Normally the executable is invoked by stage3 logstash pipeline you have in place.

example doc: (json)

{"@timestamp":"2021-03-10T00:58:21.976Z","tags":["beats_input_codec_plain_applied"],"ecs":{"version":"1.4.0"},"agent":{"hostname":"gitlab","id":"3fc83c01-53ee-4a26-b33a-fe2ec2e4dab6","type":"filebeat","version":"7.6.2","ephemeral_id":"55da0cbe-abff-4ec2-9820-624ceff143c1"},"input":{"type":"log"},"log":{"offset":0,"file":{"path":"/var/log/walt/walt.log"}},"host":{"architecture":"x86_64","name":"gitlab","containerized":false,"os":{"kernel":"4.15.0-99-generic","name":"Ubuntu","version":"18.04.4 LTS (Bionic Beaver)","codename":"bionic","platform":"ubuntu","family":"debian"},"hostname":"gitlab","id":"f1823a3617ff4004baa550a0ae8408b6"},"message":"yessir","@version":"1"}
```


## TODO list

---

add target config value (complete 03/12)

add name config value (complete 03/12)

config changes now installed //etc/logstash/conf.d (complete 03/15)

assess comparison logstash plugins to go methods (README)

evaluate debug function output

security

---
