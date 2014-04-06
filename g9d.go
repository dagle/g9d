package main

import (
	"github.com/jackyb/go-sdl2/sdl_mixer"
	"github.com/jackyb/go-sdl2/sdl"
	"code.google.com/p/go9p/p"
    "code.google.com/p/go9p/p/srv"
	"code.google.com/p/goplan9/plumb"
	"strconv"
	"fmt"
	"os"
	"bytes"
	"strings"
)

type Mode int

const (
	Play Mode = iota
	Stop
	Pause
)

type Music struct {
	srv.File
	playlist	*Playlist
	queue		*Queue
	mode		Mode
	mixer		*mix.Music
}

type Playlist struct {
	srv.File
	repeat	bool
	current int
	entries []string
}

type Queue struct {
	srv.File
	list *List
}

type List struct {
	entry	string
	next	*List
}

func Plumb(mode, msg string) {
    attr := &plumb.Attribute{
        Name:  "mode",
        Value: mode,
    }
    message := &plumb.Message{
       Src:  "mGu",
        Dst:  "music",
        Dir:  "/mnt/mGu",
        Type: "text",
        Attr: attr,
        Data: []byte(msg),
    }
    var buf bytes.Buffer
    buf.Reset()
    message.Send(&buf)
}

func (queue *Queue) Length() int {
	tmp := queue.list
	i := 0
	for ; tmp != nil; i++ {
		tmp = tmp.next
	}
	return i
}

func (music *Music) Status(cbChannel chan Mode) {
	for {
		<-cbChannel
		if music.mode == Play {
			music.Next(1)
			music.Stop()
			music.mixer = mix.LoadMUS(music.Current())
			Plumb("play", music.Current())
			music.mixer.Play(0)
		}
	}
}

func initCallback(music *Music) {
	cbChannel := make(chan Mode, 10)
	f := func () {
		cbChannel <- Play
	}
	go music.Status(cbChannel)
	mix.HookMusicFinished(f)
}

func Init() {
	sdl.Init(sdl.INIT_AUDIO)
	if (!mix.OpenAudio(44100, sdl.AUDIO_S16, mix.DEFAULT_CHANNELS, 4096)){
		fmt.Print("Could not init the mixer correctly")
	}
}

func NewMusic() *Music{
	m := new(Music)
	m.playlist = new(Playlist)
	m.queue = new(Queue)
	m.mode = Stop
	m.playlist.current = 0
	m.playlist.repeat = false
	m.queue.list = nil
	m.playlist.entries = make([]string, 0)
	return m
}

func (music *Music) Wstat(fid *srv.FFid, dir *p.Dir) error {
	return nil
}

func (music *Music) Remove(fid *srv.FFid) error {
	return nil
}

func (music *Music) Play(times int) {
	if music.mode == Stop || music.mixer == nil {
		music.mixer = mix.LoadMUS(music.Current())
	}
	Plumb("play", music.Current())
	music.mode = Play
	music.mixer.Play(0)
}

func (music *Music) Stop() {
	Plumb("stop", "")
	music.mode = Stop
	mix.HaltMusic()
	music.mixer.Free()
	music.mixer = nil
}

func (music *Music) Current() string {
	if(music.queue != nil) {
		return music.queue.list.entry
	}
	if music.playlist.current >= 0 {
		return music.playlist.entries[music.playlist.current]
	}
	return ""
}

func (music *Music) Next(steps int) {
	tmp := music.queue.list
	if steps > 0 {
		for ; tmp != nil && steps > 0; steps-- {
			tmp = tmp.next
		}
		music.queue.list = tmp
		music.playlist.current = (steps + music.playlist.current) % len(music.playlist.entries)
	} else {
		for ; steps < 0; steps++ {
			music.playlist.current--
			if(music.playlist.current < 0) {
				music.playlist.current = len(music.playlist.entries) - 1
			}
		}
	}
}

func listAdd(list **List, elems ...string) {
	tmp := *list
	for ; tmp != nil && tmp.next != nil; tmp = tmp.next {
	}
	for _,elem := range elems {
		if(tmp == nil) {
			tmp = new(List)
			*list = tmp
		}
		tmp.next = new(List)
		tmp.entry = elem
		tmp = tmp.next
	}
}

func (music *Music) Write(fid *srv.FFid, buf []byte, offset uint64) (int, error) {
	str := strings.TrimSpace(string(buf))
	if strings.HasPrefix(str,"play") && music.mode != Play {
		if len(music.playlist.entries) + music.queue.Length() > 0 {
			music.Play(0)
		}
	} else if strings.HasPrefix(str,"stop") && music.mode == Play {
		music.Stop()
	} else if strings.HasPrefix(str,"pause") {
		Plumb("play", music.Current())
		music.mode = Pause
		mix.PauseMusic()
	} else if strings.HasPrefix(str,"skip") {
		var i int
		s := strings.Split(str, " ")
		if len(s) == 1 {
			i = 1
			music.Next(1)
		} else {
			j,err := strconv.Atoi(strings.TrimSpace(s[1]))
			if err == nil {
				music.Next(j)
			}
		}
		music.Stop()
		Plumb("Skip", strconv.Itoa(i))
		music.Play(0)
	}
	return 0, nil
}

func (music *Music) Read(fid *srv.FFid, buf []byte, offset uint64) (int, error) {
	s := fmt.Sprintf("%s", music.Current())
	b := []byte(s)
	i := copy(buf, b[offset:])
	return i, nil
}

func (list *Playlist) Wstat(fid *srv.FFid, dir *p.Dir) error {
	return nil
}

func (list *Playlist) Remove(fid *srv.FFid) error {
	return nil
}

func (plist *Playlist) Write(fid *srv.FFid, buf []byte, offset uint64) (int, error) {
	s := strings.Split(string(buf), "\n")
	if offset == uint64(0) {
		plist.entries = make([]string, 0)
		plist.current = 0
	}
	plist.entries = append(plist.entries, s...)
	if offset == uint64(0) {
		Plumb("playlist set", strconv.Itoa(len(s)))
	} else {
		Plumb("Playlist added", strconv.Itoa(len(s)))
	}
	return len(buf), nil
}

func (plist *Playlist) Read(fid *srv.FFid, buf []byte, offset uint64) (int, error) {
	str := strings.Join(plist.entries, "\n")
	b := []byte(str)
	i := copy(buf, b[offset:])
	return i, nil
}

func (queue *Queue) Wstat(fid *srv.FFid, dir *p.Dir) error {
	return nil
}

func (queue *Queue) Remove(fid *srv.FFid) error {
	return nil
}

func (queue *Queue) Write(fid *srv.FFid, buf []byte, offset uint64) (int, error) {
	s := strings.Split(string(buf), "\n")
	if offset == uint64(0) {
		queue.list = nil
	}
	listAdd(&queue.list, s...)
	if offset == uint64(0) {
		Plumb("Queue set", strconv.Itoa(len(s)))
	} else {
		Plumb("Queue added", strconv.Itoa(len(s)))
	}
	return len(buf), nil
}

func (queue *Queue) Read(fid *srv.FFid, buf []byte, offset uint64) (int, error) {
	var str string
	for tmp := queue.list; tmp != nil; tmp = tmp.next {
		str += tmp.entry + "\n"
	}
	b := []byte(str)
	copy(buf, b[offset:])
	return len(b[offset:]), nil
}

func main() {
	var err error
	var music *Music
	var s *srv.Fsrv
	Init()

	user := p.OsUsers.Uid2User(os.Geteuid())
	root := new(srv.File)
	err = root.Add(nil, "/", user, nil, p.DMDIR|0777, nil)
	if err != nil {
		goto error
	}

	music = NewMusic()
    err = music.Add(root, "ctl", user, nil, 0666, music)
	if err != nil {
		goto error
	}

	err = music.playlist.Add(root, "list", user, nil, 0666, music.playlist)
	if err != nil {
		goto error
	}

	err = music.queue.Add(root, "queue", user, nil, 0666, music.queue)
	if err != nil {
		goto error
	}

	s = srv.NewFileSrv(root)
	s.Dotu = true
	s.Start(s)

	err = s.StartNetListener("tcp", "5640")
	if err != nil {
		goto error
	}

error:
	return
}
