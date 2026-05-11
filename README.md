<img src="assets/text_logo.png" alt="drawing" width="256"/></img>
# album2vid-gui
### A GUI-based tool for [npgy/album2vid](https://github.com/npgy/album2vid), powered by [Fyne](https://github.com/fyne-io/fyne).

### - - - - -
If you run a YouTube channel for music like I do, you will find this useful.
It lets you generate an output video from audio files and their respective cover art reasonably fast.

> [!NOTE] 
> Download your own FFmpeg binary sources here:
> - MacOS and Linux from https://ffmpeg.martin-riedl.de/
> - Windows x64 from https://www.gyan.dev/ffmpeg/builds/#release-builds

## Installation

The simplest way to install this tool is to download the binary for your OS on the [releases](https://github.com/HurleybirdJr/album2vid-gui/releases) page.

## Usage

`album2vid [-h] [-f] [path]`

### Flags:
`-f` or `--fast`: Enables fast mode
> [!CAUTION]
> This may cause rendering errors. You have been warned!
- > Essentially, without this flag, the program converts your input files to AAC first, and then stitches those into the output video.

### Args:  
`path`: The full path to the album folder
- > **Make sure to enclose this in quotes, for example: `"/home/HurleybirdJr/Music/CoolAlbumPath"`**

> [!IMPORTANT]
> ### Make sure you have prepared your album's files correctly. 
> They need to be <ins>ordered</ins> ( numerically or otherwise ) and the cover art needs to have a particular name ( "`cover.png`" or "`cover.jpg`" ).
> #### Here's an example folder:
> ![Album file example](https://i.imgur.com/yqjylZX.png)
> 
> Note that the tracks have their numbers in front of them, and the cover art is named "`cover.jpg`".
> This program also accepts "`cover.png`".


After it converts your audio files and cover art to a video, a file named "`out.mp4`" will appear in the same directory you ran the command on. This is your final video, you are ready to upload!  
The program also generates a "`tracklist.txt`" with timestamps for you!

> [!NOTE]
> - The output is rendered in x264 at 1080x1080.
> - Your cover art must be 1:1 aspect ratio; most are.

## File compatibility:
> [!WARNING]
> Other file types may work, but remain untested. Proceed with caution.
### Audio
`.wav`
`.mp3`
`.m4a`
`.ogg`
`.flac`
### Images
`.jpg`
`.png`

## Notes


- Nick (npgy) is not supporting macOS Intel anymore.
- Nick (npgy) is also not supporting Windows ARM64 yet

## Big thanks to:
#### Nick (npgy) for the original project! ❤️
Z from Nightride FM for help with FFMPEG  
[Alexis Masson](https://github.com/Aveheuzed) for helping refactor, organize, and simplify the codebase
