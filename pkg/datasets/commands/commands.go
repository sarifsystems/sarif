// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package commands

const Data = `
contacts/change
save a nickname for john [name:entity=john]
assign nickname bro [nickname=bro]

phone/call
Call Alice. [q=Alice]
Dial my phone number. [q=my phone number]
Call 5555. [number:number=5555]
Call John on mobile number. [name=John,phone_type=mobile]
Call last number I called. [call_type=outgoing]
Dial my last missed call. [call_type=missed]
Call John. [name=John]
Call John Smith. [name=John Smith]
Call Empire State Building. [venue_title=Empire State Building]
Call Starbucks. [venue_chain=Starbucks]
Call restaurant. [venue_type=restaurant]

phone/redial
Call Alice back. [q=Alice]
Redial my home number. [q=my number]
Redial 5555. [number=5555]
Call John again on mobile number. [phone_type=mobile]
Redial last number I called. [call_type=outgoing]
Call back my last missed call. [call_type=missed]
Call John again. [name_first=John]
Call back John Smith. [name_last=Smith]
Call Empire State Building again. [venue_title=Empire State Building]
Call Starbucks again. [venue_chain=Starbucks]
Call restaurant again. [venue_type=restaurant]

phone/answer
Answer call with speaker off. [speaker:bool=off]
Answer call with video. [video=on]

phone/show
Last number I called. [call_type=outgoing]
Show my last missed call. [call_type=missed]

phone/search
Find number for Alice. [q=Alice]
What’s my phone number. [q=my phone number]
Find John’s mobile number. [name=John,phone_type=mobile]
Do you know my home number? [phone_type=home]
Find number for John. [name_first=John]
Find number for John Smith. [name_last=Smith]

maps/search
Where is Iceland? [q=Iceland]
Find London on Google maps. [service_name=Google maps]
Open Google maps. [service_name=Google maps]
Where am I? [location=true]
What’s my current address? [address=true]
Find me on Google maps. [service_name=Google maps]

maps/navigate
Navigate to New York. [to=New York]
Navigate from New York. [from=New York]
Navigate to the nearest pub. [sort=nearest]
Navigate me to New York avoiding toll roads [road_type=free_way]
Navigate me to New York avoiding traffic [road_type=light_traffic]
Navigate me to London using the fastest way [route_type=fastest]
Navigate me to London using the shortest way [route_type=shortest]
Navigate to New York using Waze [service_name=Waze]
Directions to New York [to=New York]
Directions from New York [from=New York]
Directions to the nearest pub. [sort=nearest]
Show directions to New York using Waze [service_name=Waze]
Show me the shortest route to London [route_type=shortest]
Show me the fastest way to London [route_type=fastest]
Directions to New York avoiding toll roads [road_type=free_way]

tweet/show
Read my last tweet. [type=last]
Read my tweets. [type=all]

web/search
Search web for phones. [q=phones]
Search DuckDuckGo for phones. [service_name=DuckDuckGo]


app/open
Open angry birds. [app_name=angry birds]

app/download
Download angry birds from app store. [store_name=app store]

app/sign_in
Sign in to foursquare. [service_name=foursquare]

venue/book/hotel
Book Gloria Hotel. [venue_title=Gloria]
Book Marriott. [venue_chain=Marriott]
Book hotel in Berlin. [location=Berlin]
Book hotel nearby. [current_location=true]
Book hotel in Berlin for July 23. [date=July 23]
Book hotel in Berlin from 23 July till 25 July. [date_period=23 July till 25 July]
Book hotel for 2 persons. [adults=2]
Book hotel for 3 nights. [nights=3]
Book 5 star hotel. [stars=5]
Book hotel with free wi-fi. [venue_facility=wifi]
Reserve a hotel room equipped with air conditioner. [room_facility=air conditioning]
Book all inclusive resorts. [board_type=AI]

venue/book/restaurant
Book Red Lobster. [venue_title=Red Lobster]
Book indian restaurant. [cuisine=indian]
Book Red Lobster on Times Sq. [location=Times Sq]
Book Red Lobster for Saturday. [date=sti date]
Book Red Lobster for 7 pm. [time=sti date]
Book Red Lobster for 3 persons. [people=3]

venue/book
Book Bonsai. [q=Bonsai]
Book Zaitinya in Washington. [location=Washington]

calc/convert/currency
Convert 3 rubles in dollars. [amount=3]
Convert 3 rubles in dollars. [code=RUB]
Convert 3 rubles in dollars. [name_sg=Ruble]
Convert 3 rubles in dollars. [name_pl=Rubles]
Convert 3 rubles in dollars. [code=USD]
Convert 3 rubles in dollars. [name_sg=US Dollar]
Convert 3 rubles in dollars. [name_pl=US Dollars]

calc/tip
100 dollar plus 20% tip. [amount_without_tip=100]
100 dollar plus 20% tip. [tip_percentage=20]
100 dollar plus 30 dollar tip. [tip_amount=30]
100 dollar plus 20% tip for 5 people. [people_number=5]
100 dollar plus 20% tip. [result_full=120]
100 dollar plus 20% tip for 5 people. [result_person=24]

calc/convert
Convert 3 miles in kilometers. [amount=3]
Convert 3 miles in kilometers. [unit_code=mi]
Convert 3 miles in kilometers. [unit_name_sg=mile]
Convert 3 miles in kilometers. [unit_name_pl=miles]
Convert 3 miles in kilometers. [category=length]
Convert 3 miles in kilometers. [unit_code=km]
Convert 3 miles in kilometers. [unit_name_sg=kilometer]
Convert 3 miles in kilometers. [unit_name_pl=kilometers]
Convert 3 miles in kilometers. [category=length]
 
email/send
Email John. [recipient_name=John]
Email john@example.com [recipient_address=john@example.com]
Email John’s work address. [email_type=work]
Email John how are you. [message=how are you]
Email subject hello message how are you. [subject=hello]
Email Hello how are you? [q=Hello]

email/forward
Forward email to all my contacts [recipient=all]
Forward email to John. [recipient_name=John]
Forward email to john@example.com [recipient_address=john@example.com]
Forward email from John. [sender_name=John]
Forward email from john@example.com [sender_address=john@example.com]

email/resend
Resend email how are you to John. [message=how are you]
Resend email with subject hello. [subject=hello]

email/reply
Reply to mail from John. [sender_name=John]
Reply to john@example.com email. [sender_address=john@example.com]
Reply to John’s work address. [email_type=work]
Reply to email from John thank you. [message=thank you]
Reply to email with subject hello. [subject=hello]

email/show
Read emails from folder travel. [folder=travel]
Read new email. [unread=true]
Read mails to John. [recipient_name=John]
Read my mails to john@example.com? [recipient_address=john@example.com]
Read email from John. [sender_name=john]
Read email from john@example.com [sender_address=john@example.com]
Read emails with text how are you. [message=how are you]
Read emails with subject hello. [subject=hello]

email/check
Check email from folder travel. [folder=travel]
Any new emails? [unread=true]
Check my mails to John. [recipient_name=John]
Check mails to john@example.com? [sender_address=john@example.com]
Do I have mails from John? [sender_name=John]
Any new mails from john@example.com? [sender_address=john@example.com]
Do I have any mails with text hello? [message=hello]
Do I have any mails with subject hello? [subject=hello]

email/notify
Notify about new emails from folder travel. [folder=travel]
Notify about new emails from John. [sender_name=John]
Notify about new emails from john@example.com [sender_address=john@example.com]
Notify about new emails containing hello how are you. [message=hello how are you]
Notify about new emails. [unread=true]

email/notify/false
Don’t notify about new emails from folder travel. [folder=travel]
Don’t notify about new emails from John. [sender_name=John]
Don’t notify about new emails from john@example.com [sender_address=john@example.com]
Don’t notify about new emails containing hello how are you. [message=hello how are you]
Don’t notify about new emails. [unread=true]
 
sms/send
Message John. [recipient_name=John]
Message +7(999)555-33-22 [recipient_number=79995553322]
Message John on mobile. [phone_type=mobile]
Message John How are you? [message=How are you?]
Message Hello how are you? [q=Hello]

sms/forward
Forward sms to all my contacts [recipient=all]
Forward this text to John. [recipient_name=John]
Forward the message to +7(999)555-33-22 [recipient_number=79995553322]
Forward this text to John’s work number. [phone_type=work]
Forward John’s message. [sender_name=John]
Forward sms from +79555197810 [sender_number=+79555197810]
Resend my sms where are you. [message=where are you]

sms/reply
Reply to John’s message. [sender_name=John]
Reply to +79555597810 message. [sender_number=+79555597810]
Send a reply message to John’s mobile. [phone_type=mobile]
Reply to message from John thank you. [message=thank you]
Reply text message to google hangouts. [service=hangouts]

sms/show
Read messages from travel folder. [folder=travel]
Read messages from inbox. [folder=inbox]
Read new message. [unread=true]
Read texts to John [recipient_name=John]
Read text to +7(999)555-33-22 [recipient_number=79995553322]
Read texts to John’s mobile. [phone_type=mobile]
Read texts from John. [sender_name=John]
Read texts from +7(999)555-33-22 [sender_number=79995553322]
Read texts from John’s mobile. [phone_type=mobile]
Read messages with text Hello. [message=Hello]
Read my new viber messages. [service=viber]

sms/check
Check messages from travel folder. [folder=travel]
Are there any new messages? [unread=true]
Check my texts to John. [recipient_name=John]
Check text to +7(999)555-33-22 [recipient_number=79995553322]
Check texts to John’s mobile. [phone_type=mobile]
Check texts from John. [sender_name=John]
Check texts from +7(999)555-33-22 [sender_number=79995553322]
Check texts from John’s mobile. [phone_type=mobile]
Check messages with text Hello. [message=Hello]
Check my whatsapp messages. [service=whatsapp]

sms/notify
Notify me about new message from folder travel. [folder=travel]
Notify about new text from john [sender_name=John]
Notify about new text from +7(999)555-33-22 [sender_number=79995553322]
Notify about new text containing hello how are you. [message=hello how are you]
Notify me about new viber messages. [service=viber]
Notify me about new viber messages. [unread=true]

sms/notify/false
Don’t notify me about new message from folder travel. [folder=travel]
Don’t notify about new text from john [sender_name=John]
Don’t notify about new text from +7(999)555-33-22 [sender_number=79995553322]
Don’t notify about new text containing hello how are you. [message=hello how are you]
Don’t notify me about new viber messages. [service=viber]
Don’t notify me about new viber messages. [unread=true]

wifi/turn/on
Turn on wifi. [module=wifi]

wifi/turn/off
Turn wifi off. [module=wifi]

wifi/check
Check wifi. [module=wifi]

knowledge/query
Who built the Tower of London? [q=Tower of London]
How old is the Colosseum? [request_type=age]
What city is Eiffel Tower located in? [request_type=location in units]
IBM logotype. [q=IBM]
What is the homepage of IBM? [request_type=website]
Who is the founder of Microsoft? [request_type=founder]
Who is taller Superman or Batman? [q=Superman vs Batman]
Who is older Cristiano Ronaldo or Lionel Messi? [request_type=comparative]
What is the largest bone in the human body? [q=largest bone]
What is the smallest planet? [request_type=superlative]
Who was born on 13th of april 1961? [q=April 13 1961]
What holiday is on the 25th of December? [request_type=holiday on date]
How many hours between 14 April and 3 May? [request_type=time units between]
How far is it from Russia to Japan? [q=Russia and Japan]
How many kilometers is it from Boston to Alaska? [request_type=distance between in units]
Travel time between Moscow and London. [request_type=travel time between]
Formula for a circle area. [q=circle]
What is the area of a circle with a radius of 2 cm? [request_type=shape property to find]
What’s the diameter of a basketball? [request_type=diameter]
What is the value of e? [q=e constant]
What is the value of pi to 15 digits? [request_type=math constant value]
What is infinity divided by infinity? [request_type=infinity]
Do you know poker? [q=poker]
What is poker? [request_type=whatis]
Chess rules. [request_type=rules]
Who won the ww2? [q=World War II]
When was the First World War? [request_type=dates]
When started World War 2? [request_type=war start date]
What is Halloween? [q=Halloween]
When is Columbus day in 2015? [request_type=date in year]
How many days till Christmas? [request_type=time units until]
How old is Catholic Church? [q=Catholic Church]
How old is Catholic Church? [request_type=age]
How many people speak Portuguese? [q=Portuguese language]
English language classification. [request_type=classification]
Tell me about German language. [request_type=whatis]
Define love. [q=love]
What rhymes with the word way? [request_type=rhyme]
Do you know antonyms for good? [request_type=antonyms]
Tell me about War and Peace. [q=War and Peace]
How old is the US Constitution? [request_type=age]
When was War and Peace published? [request_type=publication date]
Tell me about Jacques Paganel. [q=Jacques Paganel]
Tell me about Jacques Paganel. [request_type=whatis]
List of books featuring Sherlock Holmes. [request_type=works featuring]
Leo Tolstoy books. [q=Leo Tolstoy]
What books Leo Tolstoy wrote? [request_type=books]
Bruce Willis filmography. [q=Bruce Willis]
Will you name some Bruce Willis movies? [request_type=movies with cast member]
What was Bruce Willis first movie? [request_type=oldest movie]
Quentin Tarantino filmography. [q=Quentin Tarantino]
What is Pulp Fiction? [q=Pulp Fiction]
Pulp Fiction actors. [request_type=cast]
What was the first Star Wars movie? [request_type=oldest movie]
Do you know Californication album? [q=Californication]
Do you know Californication album? [request_type=whatis]
Track list of Californication. [request_type=track list]
Do you know Nirvana? [q=Nirvana]
What was the second Nirvana album? [request_type=album by order]
What was the first Queen album? [request_type=oldest album]
What did Stravinsky compose? [q=Igor Stravinsky]
Stravinsky compositions. [request_type=compositions]
Do you know what Stravinsky composed? [request_type=compositions]
What is rock? [q=Rock]
Rock albums. [request_type=albums]
Tell me about rock performers. [request_type=bands]
Tell me about Stairway to Heaven. [q=Stairway To Heaven]
Who sang Stairway to Heaven? [request_type=singer]
What is Stairway to Heaven? [request_type=whatis]
What is the meaning of name Alice? [q=Alice]
What is the meaning of your name? [request_type=your name meaning]
David name meaning. [request_type=name meaning]
How far is the Sun? [q=Sun]
How many moons does the Mars have? [request_type=number of moons]
How far is the Moon right now? [request_type=current distance from Earth]
How big a bulldog gets? [q=bulldog]
What does a lion eat? [request_type=diet]
How long is a cat pregnant? [request_type=duration of pregnancy]
Guacamole recipe. [q=guacamole]
How many calories in 100 grams of guacamole? [request_type=nutrition facts in portion]
Guacamole recipe. [request_type=recipe]
GDP per capita in Venezuela. [q=Venezuela]
What is the capital of Brazil? [request_type=capital]
What state is Cupertino in? [request_type=geoobject location]
What is the speed of sound? [q=sound]
Value of planck constant. [request_type=physical constant value]
Boiling point of water in kelvins. [request_type=boiling point in units]
How big is the Bermuda Triangle? [q=Bermuda Triangle]
What is Bermuda Triangle? [request_type=whatis]
Everest height in miles. [request_type=height in units]
Color of a dandelion. [q=dandelion]
Who created the world? [request_type=creator]
Paintings of Vincent Van Gogh. [q=Vincent Van Gogh]
What paintings did Van Gogh create? [request_type=artworks]
What is Mona Lisa? [q=Mona Lisa]
Who created Mona Lisa? [request_type=painter]
Tell me about Nelson Mandela. [q=Nelson Mandela]
How tall is Barack Obama? [request_type=height]
What was Einstein’s birthday? [request_type=birth date]
Adolf Hitler became chancellor of Germany in which year? [q=Adolf Hitler]
What president was JFK? [request_type=leadership position]
Boston Marathon age. [q=Boston Marathon]
When is the next Boston Marathon? [request_type=sportsevent date]
When was the 2nd Boston Marathon? [request_type=sportsevent date by order]
Novak Djokovic current team. [q=Novak Djokovic]
Novak Djokovic current team. [request_type=current team]
Goalkeeper of Real Madrid. [q=Real Madrid]
Who is the goalkeeper of Real Madrid? [request_type=player]
What is centreboard? [q=centreboard]
Tell me about healthy food. [request_type=whatis]
When will it happen? [request_type=whenfuture]

events/find
Concerts in London. [event_type=concerts]
Chicago musical. [event_title=Chicago]
Concerts in London. [location=London]
Movies nearby. [current_location=true]
Events for today. [date=sti date]
Events this week. [date_period=startDate=sti date, endDate=sti date]
Events at 5 pm. [time=sti date]
Events this evening. [time_period=startTime=sti date, endTime=sti date]
Depeche Mode concerts. [artist=Depeche Mode]
Opera performances at Bolshoi Theatre. [venue_type=theater]
Opera performances at Bolshoi Theatre. [venue_title=Bolshoi Theatre]
Search concerts in London on eventful. [service_name=Eventful]

stocks/show
Stocks prices for Alice. [q=Alice]
Stocks prices for Twitter. [company_name=Twitter, Inc.]
Stocks prices for Twitter. [company_ticker=TWTR]
Dow shares. [index_name=dow]
Dow shares. [index_ticker=dji]
Stock price for Twitter yesterday. [date=sti date]
Find dow shares on Yahoo Finance. [service_name=Yahoo Finance]

flight/show
Check flight ba 413. [flight_number=BA413]
When will the flight from Rome depart? [departure_location=Rome]
Show me departures to New York. [arrival_location=New York]

language/set
I speak German. [request_type=user]
Speak German. [lang=German]
Speak European Portuguese. [dialect=European]
Speak German. [langCode=de]

phrases/set
When I say “Hi” you should say “Hello”. [input=Hi]
When I say “Hi” you should say “Hello”. [reply=Hello]
When I say “Hi” you should smile. [action=smile]

phrases/delete
Remove command “Hi”. [input=Hi]
Remove command when I say “Hi” you should say “Hello”. [reply=Hello]
Remove command when I say “Hi” you should smile. [action=smile]

phrases/list
Show all my commands. [all=true]

maps/search
Address of the nearest cafe [request_type=address]
Find hotel nearby [request_type=search]
Find a restaurant [venue_type=restaurant]
Find Seven Glaciers restaurant [venue_title=Seven Glaciers]
Find FamilyMart [venue_chain=FamilyMart]
Find a restaurant in new york [location=new york]
Find a restaurant nearby [sort=nearest]
lind a good restaurant [sort=best]
Find the cheapest supermarket nearby [sort=nearest, sort2=cheapest]
Find 3 stars restaurant [stars=3]
Find 5 star hotel [stars=5]
Find restaurant serving italian cuisine [cuisine=italian]
Find french restaurant [cuisine=french]
Find the place where I can eat hamburger [dish=hamburger]
Find a place to grab a snack [meal=snack]
Where I can have my breakfast [meal=breakfast]
I want to drink coffee [beverage=coffee]
Find a cafe with free wi-fi [venue_facility=wifi]
I need a hotel with free wi-fi and parking place [venue_facility=parking]
Find a hotel room equipped with air conditioner [room_facility=air conditioning]
I search for a hotel room with terrace and good view [room_facility=terrace]
Find all inclusive resorts [board_type=AI]
Find a hotel offering breakfast and dinners only [board_type=HB]
Find a restaurant on tripadvisor [service_name=tripadvisor]

maps/traffic
Any accidents nearby? [traffic_event=accidents]
Is there traffic on Main street? [location=Main street]
What’s the traffic between London and Bristol? [location_between=London and Bristol]
Any accidents nearby? [current_location=true]
Is there traffic on my way? [en_route=true]
How long is this traffic jam? [info=length]

maps/address/set
Save my home address. [type=home]
I live at 7475 Franklin Ave. [address=7475 Franklin Ave]
I work here [address=current]
Change my home address. [type=home]
What’s my home address? [type=home]
Remove my home address. [type=home]

device/toggle
Ipod off. [action=off]
Ipod on. [source=ipod]

media/search/music
Search blabla music. [q=blabla]
Look for beatles music. [artist=beatles]
Find me blackbird song. [title=blackbird]
Find white album. [album=white album]
Find rock music. [genre=rock]
Find my faves playlist. [playlist=faves]
Find some music for John. [target_name=John]
Find music for rear left seat. [target_seat=r2-left]
Find music for rear row. [target_row=r2]
Search a song on youtube. [service_name=youtube]
Search for some music on ipod. [source=ipod]

media/play/music
Play blabla music. [q=blabla]
Play artist Nightwish. [artist=Nightwish]
Play music by Nightwish. [artist=Nightwish]
Play beatles. [artist=beatles]
Play song blackbird. [title=blackbird]
Play white album. [album=white album]
Play some rock music. [genre=rock]
Play my faves playlist. [playlist=faves]
Play some music for John. [target_name=John]
Play music for rear left seat. [target_seat=r2-left]
Play music for rear row. [target_row=r2]
Play a song from youtube. [service_name=youtube]
Play some music on ipod. [source=ipod]

media/search/video
Search blabla video. [q=blabla]
Look for beatles music videos. [artist=beatles]
Find me blackbird song video. [title=blackbird]
Find some horror movies. [genre=horror]
Find some movies for John. [target_name=John]
Find movies for rear left seat. [target_seat=r2-left]
Find movies for rear row. [target_row=r2]
Search for videos on youtube. [service_name=youtube]
Search for some music videos on ipod. [source=ipod]

media/play/video
Play blabla video. [q=blabla]
Play disney cartoons. [artist=disney]
Play Lilo & Stitch. [title=Lilo & Stitch]
Play some horror films. [genre=horror]
Play some video for John. [target_name=John]
Play a movie for rear left seat. [target_seat=r2-left]
Play video for rear row. [target_row=r2]
Play a video from youtube. [service_name=youtube]
Play some video on ipod. [source=ipod]
Search blabla on radio. [q=blabla]
Find beatles on radio. [artist=beatles]
Find song blackbird on radio. [title=blackbird]

media/search/radio
Find BBC Radio 1 [station=BBC Radio 1]
Radio stations playing rock. [genre=rock]
Find any FM radio. [radio_type=fm]
Find radio station for John. [target_name=John]
Find radio for rear left seat. [target_seat=r2-left]
Find interesting radio for rear row. [target_row=r2]

media/play/radio
I need radio on car stereo. [source=car stereo]
Play blabla on radio. [q=blabla]
Play beatles on radio. [artist=beatles]
Play blackbird song on radio. [title=blackbird]
Turn on BBC Radio 1 [station=BBC Radio 1]
Turn on radio stations playing rock. [genre=rock]
Play FM radio. [radio_type=fm]
Play radio station for John. [target_name=John]
Play radio for rear left seat. [target_seat=r2-left]
Play interesting radio for rear row. [target_row=r2]
Play radio on car stereo. [source=car stereo]

media/stop/music
Stop this song. [media_type=music]
Stop music on ipod. [source=ipod]

media/stop/video
Pause the video. [media_type=video]
Pause music on ipod. [source=ipod]

media/next/radio
Next radio station. [media_type=radio]

media/next/music
Next music on ipod. [source=ipod]

media/prev/music
Previous song. [media_type=music]
Previous music on ipod. [source=ipod]

media/mute/radio
Turn off radio volume. [media_type=radio]

media/mute/music
Mute music on ipod. [source=ipod]

media/unmute/music
Unmute song. [media_type=music]
Unmute music on ipod. [source=ipod]

news/search
News about dogs. [keyword=dogs]
Fashion news. [topic=fashion]
BBC news. [source=BBCcom]
Latest news. [sort=latest]
News. [sort=top]
Radio news. [radio=true]
News about cats on Google News. [service_name=Google News]

notes/save
Save a note titled call brother. [title=call brother]
Save a note call brother. [text=call brother]
Save a note tagged call brother. [tag=call brother]
Save a note tagged call brother to evernote. [service_name=evernote]

notes/show
Get note titled call brother. [title=call brother]
Get note call brother. [text=call brother]
Get note tagged call brother. [tag=call brother]
Get note call brother from your memory. [service_name=agent_memory]

notes/list
Show all my notes. [all=true]

notes/delete
Remove note titled call brother. [title=call brother]
Remove note call brother. [text=call brother]
Remove note tagged call brother. [tag=call brother]
Remove note 1. [number=1]
Remove the last note. [number=last]
Remove all my notes. [all=true]
Remove note tagged call from evernote. [service_name=evernote]

reminder/set
Remind me at 3 PM. [time=3 PM]
Remind me in 6 hours. [duration=6 hours]
Remind about running. [summary=running]
Remind me to call Tom when I get home. [location=home]
Remind about running every day. [recurrence:const=every day]
Reminder for today. [date=today]
Set reminder to 5 pm. [time=5 pm]

reminder/list
Do you plan to remind me about running? [summary=running]
Reminders scheduled for when I come home. [location=home]
Do I have reminders for every weekend? [recurrence:const=weekend]
Show all notifications. [all=true]
Reminders scheduled for 10 October. [date=10 October]
Are you supposed to remind me about something at 7 pm? [time=7 pm]
How many notifications do I have? [size=true]

reminder/delete
Don’t remind me about meeting with John [summary=meeting with John]
Remove reminders scheduled for when I come home. [location=home]
Remove my every Wednesday reminder. [recurrence:const=Wednesday]
Remove all my reminders. [all=true]
Remove reminders set for tomorrow. [date=tomorrow]
Cancel the 5 pm reminder. [time=5 pm]

shopping/search
Buy smartphone. [q=smartphone]
I need to buy a birthday present. [occasion=birthday]
Buy glasses on fancy. [service_name=fancy]

phrases/do
Are you still there? [simplified=are you there]
Can you help me something? [simplified=can you help]
Are you eating? [simplified=do you eat]
That’s great! [simplified=good]
This is horrible! [simplified=bad]
That’s my pleasure! [simplified=you are welcome]
Sure! [simplified=yes]
Not right now. [simplified=later]
Cancel [simplified=cancel]
Give me a second. [simplified=hold on]
Tell me a secret. [simplified=secret]
That was funny! [simplified=ha ha]
I want to see you surprised. [simplified=be surprised]
Can you become sad. [simplified=be sad]
Can you wink? [simplified=wink]
See you later! [simplified=see you]
What’s up? [simplified=what is up]
Good afternoon! [simplified=hello]
Where i was born? [simplified=my birth place]
How old are you? [simplified=your age]
Are you single? [simplified=are you married]
I want to travel. [simplified=travel]
I like sport. [simplified=sport]
I like kayaking. [simplified=hobby]
Could you jump? [simplified=can you]
Am I singing? [simplified=am i]
Are you a cat? [simplified=are you]
I really like you! [simplified=i like you]
I trust you. [simplified=i believe you]
I’ll be right back. [simplified=i will be back]

lights/toggle/on
turn on the lights in the kitchen [location=kitchen]
romantic lights [type=romantic]
green light for kitchen. [color=green]
turn on the lights when I enter [condition_name=enter]
turn off the lights when I leave bathroom [location=bathroom]
turn on the lights at 6 pm [time=stiDate]
turn up the lights in the kitchen [location=kitchen]

lights/toggle/off
turn off the lights after 10 pm [time_range_start=stiDate]
turn off the lights from 10 pm till 8 am [time_range_end=stiDate]
turn off the light at 8 pm every day. [recurrence:const=every_day]
switch off all the lights. [all=true]

lights/turn/down
turn down lights for 20% [percent_change=20]
turn down lights until 40% [percent_final=40]
turn down lights for 20 points [units_change=20]
turn down lights until 40 points [units_final=40]
turn down lights when I leave [condition_name=leave]

lights/turn/up
turn up the lights when I enter bathroom [location=bathroom]
turn up the lights at 10 pm [time=stiDate]
turn down the lights after 10 pm [time_range_start=stiDate]
turn up the lights from 10 pm till 8 am [time_range_end=stiDate]
turn down the lights at 8 pm every Tuesday. [recurrence=Tuesday]

heating/toggle/off
turn off heating in the kitchen [location=kitchen]
night heating [type=night]
turn off the heating when I go out [condition_name=leave]
turn off the heating when I leave bathroom [location=bathroom]
turn off the heating at 10 pm [time=stiDate]
turn off the heating after 10 pm [time_range_start=stiDate]
turn off the heating from 10 pm till 8 am [time_range_end=stiDate]
turn off the heating at 7 am on weekends. [recurrence:const=weekend]

heating/turn/up
turn up heating in the kitchen [location=kitchen]
turn up the heating at 6 am [time=stiDate]

heating/turn/down
turn down heating for 20% [percent_change=20]
turn down heating until 40% [percent_final=40]
turn down heating for 20 degrees [units_change=20]
turn down heating until 40 degrees [units_final=40]
turn down the heating when I go out [condition_name=leave]
turn down the heating when I leave bathroom [location=bathroom]
turn down the heating after 10 pm [time_range_start=stiDate]
turn down the heating from 10 pm till 8 am [time_range_end=stiDate]
turn down the heating at 7 am on weekends. [recurrence:const=weekend]

lock/lock
lock the door [type=door]
lock the front door [location=front]
lock the door when I go out [condition_name=leave]
lock the door at 10 pm [time=stiDate]
lock the door after 10 pm [time_range_start=stiDate]
lock the door from 10 pm till 8 am [time_range_end=stiDate]
make sure the door is locked at 7 pm every day. [recurrence:const=every day]

lock/unlock
unlock the window when I leave bathroom [location=bathroom]

lock/close
close the door [type=door]
close the front door [location=front]
close the door when I go out [condition_name=leave]
close the door after 10 pm [time_range_start=stiDate]
close the door from 10 pm till 8 am [time_range_end=stiDate]
make sure the door is closed at 7 pm every day. [recurrence:const=every day]

lock/open
open the window when I leave bathroom [location=bathroom]
open the door at 7 am [time=stiDate]

appliance/toggle/on
turn on TV [appliance_name=TV]
turn on TV in the kitchen [location=kitchen]
 
appliance/toggle/off
turn off TV when I leave [condition_name=leave]
turn off TV when I leave kitchen [location=kitchen]
turn off TV at 10 pm [time=stiDate]
turn off TV after 10 pm [time_range_start=stiDate]
turn off TV from 10 pm till 8 am [time_range_end=stiDate]
turn the computer off at 7 pm on weekends. [recurrence:const=weekend]

social/status/faceboo
Update facebook beautiful weather today [text=beautiful weather today]

social/delete
Delete my facebook post hello [text=hello]

social/notify
Notify about comments. [request_type=comments]
Notify about comments. [action=on]

social/notify/false
Do not notify about comments. [action=off]

social/show
Read my last facebook post. [type=last]
Read my updates on facebook. [type=all]

social/post
Tweet: Wonderful weekend it was! [text=Wonderful weekend it was!]
Remove my twitter status Hohoho! [text=Hohoho!]

social/notify
Notify about replies on Twitter. [request_type=replies]
Notify about replies on Twitter. [action=on]
Do not notify about replies on Twitter. [action=off]

sports
How did knicks play yesterday? [date=yesterday]
How did knicks do the last 3 games? [numberofgames=3]
Basketball standings. [discipline=basketball]
NBA standings. [league=NBA]
Scores for New York Knicks. [name_team_1=New York Knicks]
New York Knicks vs Los Angeles Lakers. [name_team_2=Los Angeles Lakers]

alarm/create
Set alarm for 6 am. [time=6 am]
Wake me up at 6 AM. [time=6 am]
Wake me up in 8 hours. [duration=8 hours]
Wake me up in the morning. [time_period=morning]
Set alarm for tomorrow. [date=tomorrow]
Set alarm for every Friday. [recurrence=Friday]
Set alarm for every day. [recurrence:const=every day]

alarm/list
Check my alarm for 6 am. [time=6 am]
Did you set alarm to wake me up in the morning? [time_period=morning]
Check alarm for tomorrow. [date=tomorrow]
Check alarm for every Friday. [recurrence=Friday]
Check every day alarm. [recurrence:const=every day]
Check all my alarms. [all=true]

alarm/change
Change my 6 am alarm to 7 am. [time=7 am]
Change my 6 am alarm. [oldTime=6 am]
Change my 6 pm alarm to morning. [time_period=morning]
Change my morning alarm. [oldTime_period=morning]
Reschedule my today alarm for tomorrow. [date=tomorrow]
Reschedule my today alarm. [oldDate=today]
Change every Friday alarm to every weekday. [recurrence=weekday]
Reschedule my every day alarm to every Monday. [recurrence=Monday]
Change every Friday alarm. [oldRecurrence=Friday]
Reschedule my every day alarm. [oldRecurrence=day]

alarm/delete
Remove my alarm for 6 am. [time=6 am]
Remove my morning alarm. [time_period=morning]
Remove alarm for tomorrow. [date=tomorrow]
Remove alarm for every Friday. [recurrence=Friday]
Remove every day alarm. [recurrence=day]
Remove all the alarms. [all=true]

calc/date
Convert Moscow time. [location_1=Moscow]
Convert Moscow time in Tokyo. [location_2=Tokyo]
What time is it in Rome, when it’s 3 pm in Tokyo. [time=StiDate]
What week day is it? [request_type=dayofweek]
Date on monday. [request_type=date]
Date on monday. [date=monday]
Tell me the date.
Is today 26 of May? [isDate=26 of May]
Is tomorrow Monday? [dayofweek=Monday]
Is it Friday or Saturday today? [dayofweek_1=Friday]
Is it Friday or Saturday today? [dayofweek_1=Saturday]
Is it February? [month=February]
Is it 2014 year? [year=2014]
Is it 21st century? [century=21]
Date in Japan. [location=Japan]
18:03 minus 7 minutes. [time_1=StiDate]
18:03 minus 7 minutes. [time_2=7]
Today plus 3 weeks. [time_1=StiDate]
Today plus 3 weeks. [time_2=3]
18:03 + 7 minutes. [math=plus]
Today plus 3 weeks. [unit=week]
Time in Jerusalem. [location=Jerusalem]
Time difference between London and Moscow. [location_1=London]
Time difference between London and Moscow. [location_2=Moscow]
Tokyo time zone. [location=Tokyo]
My time zone. [location=current]

timer/set
Set timer for 8 am. [time=StiDate]
Set timer for 2 hours 8 minutes. [hours=2]
Set timer for 2 hours 8 minutes. [minutes=8]
Set timer for 19 seconds. [seconds=19]

translate
Translate “hello”. [text=hello]
Translate “привет” from Russian. [lang=Russian]
Translate “привет” from Russian. [langCode=ru]
Translate “привет” to English. [lang=English]
Translate “привет” to English. [langCode=en]

tv/list
What’s on channel 5? [channel=5]
What’s on tv today? [date=today]
What’s on tv at 5? [time=5]
Is Bruce Willis on tv today? [person=Bruce Willis]
When does Big Bang Theory start? [show=Big Bang Theory]
Are the Cincinnati Bengals on television this week? [team=Cincinnati Bengals]
When does the movie Goodfellas come back on TV? [movie=Goodfellas]

weather/show
Weather in London. [request_type=explicit]
Is it raining outside? [request_type=condition]
Weather in London. [location=London]
Weather between London and Moscow. [location_between=London and Moscow]
Weather here. [current_location=true]
Weather now. [current_time=true]
Weather in london at 5. [time=5]
Weather in london this evening. [time_period=this evening]
Weather in london tomorrow. [date=tomorrow]
Weather in london this week. [date_period=this week]

weather/query
Is it cold outside? [condition_temperature=cold]
Is it raining outside? [condition_description=rain]
Should i take umbrella today? [outfit=rain]
Can i go climbing mt. Fuji next week? [activity=climbing]

browser/open
Open website google. [q=google]
Open twitter. [website=twitter]
Open twitter dot com. [domain=com]

image/search
Show kittens. [q=kittens]
Show kittens on 500px. [service=500px]

timetracker/start
Start tracking time.
Track time for work. [category=work]
Track time for work: issue 57. [category=work,text=issue 57]
Track time: issue 57. [text=issue:57]

timetracker/stop
Stop tracking time.
Stop tracking time for work. [category=work]
Stop tracking time for work: issue 57. [category=work,text=issue 57]
Stop tracking time: issue 57. [text=issue:57]

timetracker/list
List timesheet.
List tracked time.

event/new
Record that something happened. [text=something happened]
Note that something happened. [text=something happened]

event/show
Get last event.
Show last event with action lights/off. [action=lights/off]
Show last event from lights/off. [action=lights/off]

event/list
List events with action lights. [action=lights]
List events from timetracker. [action=timetracker]
Find events with action coffee. [action=coffee]

push/text
Push this text. [text=this text]
Push pass123 to phone. [text=pass123,device=phone]

location/show
Get last location.
Show last location.
Show last location with address Berlin. [address=Berlin]
Show last location at Baker Street. [address=Baker Street]
When did I last visit Berlin? [address=Berlin]
`
