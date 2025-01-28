# About Video Streaming in Browser

## HTML5 Video tag

[HTML5 Video tag](https://developer.mozilla.org/en-US/docs/Web/HTML/Element/video)

### Subtitles

```
<video controls src="video.webm">
  <track default kind="captions" src="captions.vtt" />
</video>
```

#### WVTT

[WVTT Format](https://developer.mozilla.org/en-US/docs/Web/API/WebVTT_API/Web_Video_Text_Tracks_Format)

[WVTT API - Javascript](https://developer.mozilla.org/en-US/docs/Web/API/WebVTT_API)

```
 WEBVTT

NOTE This is a multi-line note block.
These are used for comments by the author
Two cue blocks are defined below.

00:01.000 --> 00:04.000
Never drink liquid nitrogen.

00:05.000 --> 00:09.000
Because:
- It will perforate your stomach.
- You could die.

00:09:00 --> 00:10:00
[whispering] What's that off in the distance?
```
WEBVTT

NOTE This is a multi-line note block.
These are used for comments by the author
Two cue blocks are defined below.

00:01.000 --> 00:04.000
Never drink liquid nitrogen.

00:05.000 --> 00:09.000
Because:
- It will perforate your stomach.
- You could die.
```

#### Style
```
WEBVTT

STYLE
::cue {
  background-image: linear-gradient(to bottom, dimgray, lightgray);
  color: papayawhip;
}
/* Style blocks cannot use blank lines nor "dash dash greater than" */
```


## Authentication in video content

### Cookies
- Nothing special to do 
with every http request the browser will automatically sent the cookies


