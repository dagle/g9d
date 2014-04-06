g9d
===

An alternative to mpd designed like m9u. It uses a filesystem for controlling it.

/ctl for controlling (or reading for getting the current song)
Valid commands are: play, stop, pause, skip (int).

/list for storing a playlist
Writing into the list sets or adds objects to the playlist. The lists wraps around
in the end.

/queue for storing a queue
Queue is just like a list but each item is just played once and then removed.

Why are you using sdl atm? Because it's simple and works for now, most likely I will
change it to something with lesser dependencies and more features.
Like stream-support, spotify, etc (this should be done with some dynamic loading though).

g9d doesn't care about tags, location, etc. All it needs is that the files is acceble 
from the current namespace. So a tool (like ncmpcc) would need to read the tags from the
music location and just send the URI to the /list. A tool like mpc isn't needed since the
standard unix tools are sufficient. Though a filesystem handling tags and such might aid
you.

The other difference between m9u and g9d (except it actually plays stuff and is written in go)
is that it uses plumber for events. So you must have an instance of plumber running, if you
want events.

To install g9d you need to install the following:
plan9fromuserspace
(golang as build dep)
golang:
	sdl
	sdl-mixer (my repo, because I use a callback not supported by most)
	go9p
	goplan9/plumb

go install "github.com/dagle/go-sdl2/sdl_mixer"
go install "github.com/jackyb/go-sdl2/sdl"
go install "code.google.com/p/go9p/p"
go install "code.google.com/p/go9p/p/srv"
go install "code.google.com/p/goplan9/plumb"

The idea is that this project should never be bigger than 1000 lines of code.
The project is young and over heavy development and might change and have a lot 
of bugs atm.
