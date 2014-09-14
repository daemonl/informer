Informer
=======

Informer runs basic checks for web things.

The config file is XML (yes XML, it's just - good for this), containing the sites and servers to be checked.

A <site> is a website, or api, a set of tests for text based urls.

A <server> is an OS (or container) running somewhere.

Each <site> and <server> can have any number of <email> and <api> tags which will be emailed and called respecively for any checks which fail, or errors.

One <smtp> tag configures settings for sending emails to the various addresses.

Any number of <api name="..."> tags define API urls to call. See http://prowlapp.com/ for a very handy iOS Push app you could use.

Server Checks
-------------

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

HTTPS Client Auth
-----------------

Add a custom client tag to a site definition to use a client key and certificate

```
<client>
	<cert>/var/lib/keys/test_cert.pem</cert>
	<key>/var/lib/keys/test_key.pem</key>
</client>
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


