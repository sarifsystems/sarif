Roadmap
=======

1. [x] initial architecture
2. [x] deploy first long running services
3. [ ] initial desktop/Android clients to make the whole thing useful
4. [ ] web dashboard
5. [ ] rich media GUI clients

To Do
=======

* documentation
* test coverage
* services:
	- hostscan
		+ [x] automatic scanning
		+ [x] simple query
	- location
		+ [x] working save
		+ [x] working retrieval "when did i last visit"
		+ [x] geofencing: create zones
		+ [x] geofencing: publish events on zone/enter leave
		+ [ ] geofencing: list zones
	- scheduler
		+ [x] working prototype
		+ [x] parse 'reply' messages
	- selfspy
		+ [ ] import option for sqlite file
	- contacts
		+ [ ] vcf reading support
		+ [ ] storage
		+ [ ] query
		+ [ ] sync
	- pkgtrack
		+ [ ] store
		+ [ ] automatic status change messages
	- knowledge
		+ [ ] parsing google
		+ [ ] parsing wolfram alpha
	- natural
		+ [x] more complex parsing
		+ [ ] dynamic changes
	- event stream
		+ [x] store/query
		+ [ ] grouping of all updates (selfspy, music tracks, location, ..)
		+ [ ] daily digest + graphs
* clients:
	- private web interface
		+ [ ] mockup
		+ [ ] setup framework
		+ [ ] setup simple graphs with D3
	- desktop client
		+ [ ] notification support
		+ [ ] push interface
		+ [ ] chat
		+ [ ] global shortcut
		+ [ ] mpd integration
	- android client:
		+ [ ] notification support
		+ [ ] push interface
		+ [ ] chat
		+ [ ] publish location
