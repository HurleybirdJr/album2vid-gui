# album2vid-gui

### A GUI-based tool for npgy/album2vid, made in Python.

This fork provides an alternative GUI-based tool for album2vid.

It lets you generate a video output, from the audio files and cover art, of an album reasonably fast.

## Installation
~~The simplest way to install this tool is to download the binary for your OS on the [releases](https://github.com/npgy/album2vid/releases) page.~~

~~You can also choose to just clone the python script and run it manually with your own binary of Ffmpeg.~~

~~This should work on any OS.~~

#### Coming Soon
##### (*There are currently no pre-built binaries yet*)

## Usage
First, make sure you have prepared your album's files correctly. 
They need to be ordered numerically with the cover art named appropriately.

### Audio file compatability:
```.wav```
```.mp3```
```.m4a```
```.ogg```
```.flac```

#### Here's an example folder:  

![Album file example](https://i.imgur.com/yqjylZX.png)

Tracks should have their numbers in front of them, and the cover art is named "cover.jpg".  
**Note:** This program also accepts "cover.png".

Once you open the program, it will ask for the directory of your album's files. You can paste in the path url or just run the program inside that folder and hit enter.  
After it converts your audio files and cover art to a video, a file named "out.mp4" will appear. This is your final video, you are ready to upload!  
In addition, this program also generates a tracklist with timestamps for you! It will output to the file "tracklist.txt"


## Notes

Some things to note are that this renders in x264 and 1080x1080. 

Your cover art must also be 1:1 aspect ratio; most are.  

## TO-DO List


	<img src="https://img.shields.io/badge/STATUS-ready-success?style=flat-square">

## Credit
### [album2vid](https://github.com/npgy/album2vid) by [@npgy](https://github.com/npgy)
