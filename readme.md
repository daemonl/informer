Informer
=======

Informer runs basic checks for web things, then exits.

It is useful as either a cron job, or after 'changing things' to make sure everything still works.

The config files are XML (yes XML, it's just - good for this), containing the sites and servers to be checked.

Multiple config files are combined, all sections can be repeated except for SMTP.

A <site> is a website, or api, a set of tests for text based urls.

A <server> is an OS (or container) running somewhere.

Each <site> and <server> can have any number of <email> and <api> tags which will be emailed and called respecively for any checks which fail, or errors.

One <smtp> tag configures settings for sending emails to the various addresses.

Any number of <api name="..."> tags define API urls to call. See http://prowlapp.com/ for a very handy iOS Push app you could use.

Overall Format
--------------
```
<?xml version="1.0" encoding="UTF-8"?>
<informer>
	<site name="Example">	
		<email address="test@example.com"/>
		<email address="test2@example.com"/>
		<redirect>
			<from>http://example.com/</from>
			<to>http://www.example.com/</to>
		</redirect>
		<search url="http://www.example.com/">
			<string>Example</string>
			<string>Example2</string>
		</search>
	</site>
</informer>
```


Site Checks
-----------

A site file is
```
<site name="GitHub">
	...
</site>
```

### Search

Search for strings in the given url.

```
<search url="https://github.com">
	<string>GitHub</string>
	<string>UA-1234567-1</string>
</search>
```

### Redirect

Check that a url redirects to another

```
<redirect>
	<from>http://github.com</from>
	<to>https://github.com</to>
</redirect>
```

Client Config
-------------
The following work for Search, Redirect and Data checks


```
	<cookie name="auth_token">
		testtoken
	</cookie>
	<header name="Accept">application/json</header>
	
	<client insecure="true" timeout="10">
		<cert>/var/lib/keys/test_cert.pem</cert>
		<key>/var/lib/keys/test_key.pem</key>
	</client>
```
### Self Signed Certificates

```
inseucre='true'
```

Data - Logs
-----------
Checks a log, where lines begin with the date, for activity in the last [quietPeriod].
Date and Period formats use the golang time package

http://golang.org/pkg/time/

Note that currently, the parsing assumes a fixed width for the date / format - the first len(format) characters are parsed.

URL and client settings as per search

```
<data name='log-check'>
	<log url="http://example.com/log.log">
		<format>2006-01-02 17:05:06</format>
		<quietPeriod>1h</quietPeriod>
	</log>
</data>
```


Data - JSON
-----------
Checks for values in a JSON file.

URL and client settings as per search.

Keys are walked from the root.

time - Parses a time according to format, and makes sure it is at most [age] old.
A format of UNIX is a timestamp.

number - Compares the number in the key.
max, min, eq, neq
All are parsed as float64s, so eq and neq may behave poorly.

Future:
- array count
- string compare
- time compare properly

```
<data name='json-check'>
	<json url="http://example.com/stat.json">
		<time key="meta.reported">
			<format>UNIX</format>
			<age>1h10m</age>
		</time>
		<number key="errors.count">
			<max>20</max>
			<min>1</min>
			<eq>15</eq>
			<neq>19</neq>
		</number>
	</json>
</data>
```

Groups
------

The default runs only elements in the root object.

Specify a '--group [name]' flag to run only the group with the given name. 

Useful to separate hourly checks from daily checks.

'--group all' runs everything


```
<informer>
	<site .../>
	<group name='daily'>
		<site .../>
		<data .../>
	</group>
</informer>
```

SMTP Config
-----------

```
<smtp>
	<fromAddress>test@test.com</fromAddress>
	<password>secret</password>
	<serverAddress>smtp.gmail.com</serverAddress>
	<serverPort>587</serverPort>
	<username>informer@test.com</username>
</smtp>
```

API Config
----------

```
<api name="prowl-dw">
	<url>https://api.prowlapp.com/publicapi/add</url>
	<method>POSTFORM</method>
	<postval key="apikey">0a0a0a0a0a0a0a</postval>
	<postval key="application">Informer Fail</postval>
	<postval key="event">#title</postval>
	<postval key="description">#body</postval>
</api>
```

Server Checks
-------------

Server checks are not fully implemented yet

At this stage, only disks are checked for servers.

```
<server hostname="localhost">
	<email address="test@test.com"/>
	<email address="example+informer-localhost@gmail.com"/>
	<api name="prowl"/>
	<disk filesystem="/dev/xda1">
		<minBytes>30000000</minBytes>
	</disk>
	<disk filesystem="/dev/xda2">
		<minPercent>10</minPercent>
	</disk>
</server>
```

- minBytes: check that there are more than x bytes free on the disk
- minPercent: check that there is more than x% of the disk free.


