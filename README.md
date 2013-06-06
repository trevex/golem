golem _v0.4.0_
================================
A lightweight extendable Go WebSocket-framework with [client library](https://github.com/trevex/golem_client). 

License
-------------------------
Golem is available under the  [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0.html)

Installation
-------------------------
```
go get github.com/trevex/golem
```

Client
-------------------------
The client is still in work, but is already used in the examples and is available [here](https://github.com/trevex/golem_client).

Documentation
-------------------------
The documentation is provided via [godoc](http://godoc.org/github.com/trevex/golem).

Wiki & Tutorials
-------------------------
More informations and insights can be found on the [wiki page](https://github.com/trevex/golem/wiki) along with a tutorial series to learn how to use golem.

Examples
-------------------------
Several examples are available in the [example repository](https://github.com/trevex/golem_examples). To use them simply checkout the
repository and make sure you installed (go get) golem before. A more detailed guide on how
to use them is located in their repository.

History
-------------------------
* _v0.1.0_ 
  * Basic API layout and documentation
* _v0.2.0_ 
  * Evented communication system and routing
  * Basic room implementation (lobbies renamed to rooms for clarity)
* _v0.3.0_ 
  * Protocol extensions through Parsers
  * Room manager for collections of rooms
* _v0.4.0_ 
  * Protocol interchangable
  * Several bugfixes
  * Client up-to-date

Special thanks
-------------------------
* [Gary Burd](http://gary.beagledreams.com/) (for the great WebSocket protocol implementation and insights through his examples)
* [Andrew Gallant](http://burntsushi.net/) (for help on golang-nuts mailing list)
* [Kortschak](https://github.com/kortschak) (for help on golang-nuts mailing list)

TODO
-------------------------
* Verbose and configurable logging
* Testing
