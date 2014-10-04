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
* spec:
	- media support
		+ [ ] define basic format of rich media messages
		+ [ ] schema.org support
* services:
	- hostscan: track devices in home network via nmap
		+ [x] automatic scanning
		+ [x] simple query
	- location: handle user phone location tracking
		+ [x] working save
		+ [x] working retrieval "when did i last visit"
		+ [x] geofencing: create zones
		+ [x] geofencing: publish events on zone/enter leave
		+ [ ] geofencing: list zones
		+ [ ] last locations: return with active geofence info
		+ [ ] last locations: filter by geofence
		+ [ ] digest: walking, cycling, time spent at geofence
		+ [ ] digest: identify static locations, ask for check in
	- scheduler: send messages on specific conditions
		+ [x] working prototype
		+ [x] parse 'reply' messages
	- selfspy: manage desktop logs from selfspy
		+ [ ] import option for sqlite file
		+ [ ] dynamic import
		+ [ ] dynamic event generation (for daily digest)
		+ [ ] query support
	- contacts
		+ [ ] vcf reading support
		+ [ ] storage / query
		+ [ ] sync with carddav server
		+ [ ] relationships between contacts, groups
	- calendar
		+ [ ] ics reading support
		+ [ ] storage / query
		+ [ ] sync with caldav server
	- pkgtrack: handle status of shipment tracking numbers
		+ [ ] store
		+ [ ] automatic status change messages
	- mailfiter: scan and process incoming emails
		+ [ ] scan incoming mails
		+ [ ] save package tracking numbers
		+ [ ] parse semantic mails (schema.org, "Google Now Cards")
	- wishlist: track future film releases and other media
		+ [ ] generic things
		+ [ ] movie queue: initial setup and retrieving
		+ [ ] movie queue: check cinema / dvd / pre / torrent release dates
	- knowledge: answer questions by querying knowledge providers
		+ [x] parse google
		+ [x] parse wolfram alpha
	- natural: handle natural user text input
		+ [x] more complex parsing
		+ [ ] template support (e.g. via jq/gojee)
		+ [ ] dynamic changes
	- events: manage a general stream of events and generate daily/weekly digests
		+ [x] store/query
		+ [x] store with geofence
		+ [ ] grouping of all updates (selfspy, music tracks, location, ..)
		+ [ ] daily digest + graphs
	- daily/weekly digest:
		+ [ ] mail with stats
		+ [ ] simple graphs
		+ [ ] daily review (tag static locations -> check in)
	- weather: fetch current weather info and forecast
		+ [ ] query
		+ [ ] providers: openweathermap, ...
	- files: provide storage for smaller files (max mail attachment size)
		+ [ ] upload/retrieve files
		+ [ ] metadata support
		+ [ ] image recognition / tagging
	- renderer: generates images from messages for display in clients with missing features
		+ [ ] render dataseries as scatterplot
		+ [ ] render GeoJSON tracks and map tiles
		+ [ ] render LaTeX / MathML formulas
	- tracker:
		+ [ ] standardized questions, natural support
		+ [ ] mood tracking, randomized, verbal and numeric
		+ [ ] activity tracking
		+ [ ] meal tracking (serving/energy/carbs/fat/proteins), daily review
	- qr code generator
	- thing plusplus
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
