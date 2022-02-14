# album2vid-gui

### A GUI-based tool for npgy/album2vid, made in Python.

This fork provides an alternative GUI-based tool for album2vid.  
It lets you generate a video output, from the audio files and cover art, of an album reasonably fast.  
  
## TO-DO List
![TO-DO Key Success](https://img.shields.io/badge/STATUS-ready-success?style=flat-square)
![TO-DO Key In Progress](https://img.shields.io/badge/STATUS-in%20progress-important?style=flat-square)
![TO-DO Key Not Started](https://img.shields.io/badge/STATUS-waiting-critical?style=flat-square)

| Feature | Details | Status |
| - | - | - |
| `Update to @ngpy v1.3.1` | Backend ffmpeg code update from @npgy | ![TO-DO Key In Progress](https://img.shields.io/badge/STATUS-in%20progress-important?style=flat-square) |
| `Enabling custom render sizes` | Provides a GUI menu for custom video render sizes | ![TO-DO Key Not Started](https://img.shields.io/badge/STATUS-waiting-critical?style=flat-square) |
| `Extra GUI options (fast mode)` | Fast-mode skips initial encoding for rare ffmpeg bug | ![TO-DO Key Not Started](https://img.shields.io/badge/STATUS-waiting-critical?style=flat-square) |

## Installation
#### Coming Soon
##### (*There are currently no pre-built binaries yet*)

## Usage
### Audio file compatability:
**```.wav```**
**```.mp3```**
**```.m4a```**
**```.ogg```**
**```.flac```**  

First, make sure you have prepared your album's files correctly.  
Tracks should be ordered numerically with the cover art named appropriately.  

#### Here's an example folder:  

![Album file example](https://i.imgur.com/yqjylZX.png)

Tracks should have their numbers in front of them, and the cover art named **"cover.jpg"**.  
 **Note:** This program also accepts **"cover.png"**.  

![GUI preview](https://imgur.com/tjN0DcZ.png)

Once you open the program, click **`Choose Folder`** to load the album directory of your album's files.

By default, before the folder is loaded, there is an inital encoding phase for the audio.  
This is done to prevent a rare ffmpeg bug.  
  
#### (*A "fast-mode" option is coming soon, check [TO-DO List](https://github.com/HurleybirdJr/album2vid-gui/edit/beta/README.md#to-do-list)*)  
  
After it converts your audio files, you can preview the tracks it has found and the cover art above it.  
Once ready, click **`Render`**, and the render process will begin.  
  
When the render is complete, a file named **"out.mp4"** will appear in the same directory, and timestamps will be generated in **"tracklist.txt"**.  
  
### Now you are ready to upload!  


## Notes

Some things to note are that this renders in **x264** and **1080x1080**.  
Your cover art must also be **1:1 aspect ratio**; most are.  

#### (*Custom render options are coming soon, check [TO-DO List](https://github.com/HurleybirdJr/album2vid-gui/edit/beta/README.md#to-do-list)*)  

## Credit
### [album2vid](https://github.com/npgy/album2vid) by [@npgy](https://github.com/npgy)
