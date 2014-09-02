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
* core:
	- mux
		+ [ ] fix duplicated messages on subscription
* services:
	- hostscan
		+ [x] automatic scanning
		+ [x] simple query
	- location
		+ [x] working save
		+ [x] working retrieval "when did i last visit"
		+ [ ] geofencing: create/list zones
		+ [ ] geofencing: publish events on zone/enter leave
	- scheduler
		+ [ ] working prototype
		+ [ ] parse 'send' messages
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
		+ [ ] more complex parsing
		+ [ ] dynamic changes
	- event stream
		+ [ ] grouping of all updates (selfspy, music tracks, location, ..)
		+ [ ] query
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
