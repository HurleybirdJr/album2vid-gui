# AUTHOR: Nicholas Preston (npgy)
# CO-AUTHOR (GUI): Will Hurley (HurleybirdJr)

from glob import glob
import os
import subprocess
import mutagen
import time
import argparse
from sys import exit, platform
from tkinter import *
from tkinter import filedialog
from tkinter import ttk
from PIL import Image, ImageTk

parser = argparse.ArgumentParser(description="A command line tool for generating videos from albums/tracks")
parser.add_argument('-f', '--fast', action='store_true', help="Enables fast mode, may cause rendering errors")
parser.add_argument("path", nargs="?", default="", help="The full path to the album's folder")
args = parser.parse_args()


def get_runtime(filename):
    """Returns the runtime of a file

    ;param filename: the file's full path
    ;return: the file's runtime as a float
    """

    raw_info = mutagen.File(filename).info.pprint()
    runtime = float(raw_info.rsplit(', ', 1)[1].split(' ')[0])
    return runtime


def get_shortname(filename):
    """Returns just the file's name

    ;param filename: the file's full path
    ;return: the file's name
    """
    return filename.split('/')[-1].split('\\')[-1].rsplit('.', 1)[0]


def get_timestamp(seconds):
    """Returns the formatted timestamp

    ;param seconds: the length of time in seconds
    ;return: the formatted timestamp
    """
    return time.strftime("%H:%M:%S", time.gmtime(seconds))


def throw_error(text):
    print("ERROR: " + text)
    exit()


def cleanup():
    """
    Cleans up temporary files that may have been generated
    """

    try:
        # Remove unneeded files list file
        os.remove(directory + "files.txt")

        # Remove temp dir files
        for temp_file in files:
            os.remove(f"{temp_dir}/{get_shortname(temp_file)}.m4a")
    except FileNotFoundError:
        pass
    try:
        # Remove temp dir
        os.rmdir(temp_dir)
    except FileNotFoundError:
        pass


# FFMPEG binary location
ffmpeg = "ffmpeg"

# Fix command for linux systems
if platform == "linux":
    ffmpeg = "./" + ffmpeg

print("Welcome to album2vid!")

# # # GUI # # #
root = Tk()

# # # GUI attributes
root.title("album2vid")
root.resizable(False, False)
root["bg"] = "#1c1c1c"
root.iconbitmap("cool.ico")

gui_width = 1000
gui_height = 700

# Locate display center from display dimensions
center_x = int(root.winfo_screenwidth() / 2 - gui_width / 2)
center_y = int(root.winfo_screenheight() / 2 - gui_height / 2)

# Position GUI to display center
root.geometry(f'{gui_width}x{gui_height}+{center_x}+{center_y}')


# # # GUI Functions
def getFolder():
    global cmd
    global directory
    global files
    global temp_dir

    directory = filedialog.askdirectory()

    temp_dir = directory + "/.temp"

    # Verify directory is valid
    if not (os.path.isdir(directory) or directory == ""):
        throw_error("The directory could not be found")

    # Add trailing slash for expanding filenames
    try:
        if directory[len(directory) - 1] != "/":
            directory += "/"
    except IndexError:
        pass

    # Find all audio files in the directory
    extensions = ('wav', 'mp3', 'm4a', 'ogg', 'flac')
    files = []
    extract_tracks = []
    for ext in extensions:
        extract_tracks.append(glob("*." + ext))
        files.extend(glob(directory + "*." + ext))
    if len(files) == 0:
        throw_error("The audio files could not be found")

    # Find cover art
    cover = []
    cover.extend(glob(directory + "cover.jpg"))
    cover.extend(glob(directory + "cover.png"))
    try:
        cover = cover[0]
    except IndexError:
        throw_error("The cover photo could not be found")

    # Sort the list of files
    files.sort()
    extract_tracks = list(filter(None, extract_tracks))
    track_list = []
    for i in range(0, len(extract_tracks[0])):
        track_list.append(extract_tracks[0][i])
    track_list.sort()
    print(track_list)

    # Clean up temp files in case program was exited abruptly on last run
    cleanup()

    # Calculate track-list with timestamps
    with open(directory + "tracklist.txt", "w") as f:
        curr_time = 0.0
        for file in files:
            f.write(get_shortname(file) + " -- " + get_timestamp(curr_time) + "\n")
            curr_time += get_runtime(file)

    # Write all audio files to a temporary text document for ffmpeg
    with open(directory + "files.txt", "w") as f:
        for file in files:
            file = f"{temp_dir}/{get_shortname(file)}.m4a"
            # This part ensures that any apostrophes are escaped
            file = file.split("'")
            if len(file) > 1:
                file = "'\\''".join(file)
            else:
                file = file[0]

            # Write the file line
            f.write("file '" + file + "'\n")

    album2_image_file = ImageTk.PhotoImage(Image.open(cover).resize((300, 300)))
    album_image_label.configure(image=album2_image_file)
    album_image_label.image = album2_image_file

    tracks = StringVar(value=track_list)
    print(tracks)
    album_track_list_preview = Listbox(root, listvariable=tracks, height=30, width=50)
    album_track_list_preview.place(x=0, y=300)

    # First pass to encode audio (this avoids errors in the final render)
    os.mkdir(temp_dir)
    for file in files:
        first_pass_cmd = f"""{ffmpeg} -i "{file}" -map 0 -map -v? -map V? -acodec aac -b:a 320k "{temp_dir}/
        {get_shortname(file)}.m4a" """
        subprocess.run(first_pass_cmd)

    go_button.configure(bg="#a9ff4d")

    cmd = f"{ffmpeg} -y -loop 1 -framerate 1 -i {cover} -f concat -safe 0 -i {directory}files.txt -tune stillimage " \
          f"-shortest -fflags +shortest -max_interleave_delta 100M -vf format=yuv420p -s 1080x1080 -b:a" \
          f" 320k {directory}out.mp4 "

    return cmd


def startRender():
    # Execute FFMPEG command and delete temporary file
    subprocess.run(cmd)

    # Clean up temp files
    os.remove(directory + "files.txt")

    # Remove temp dir
    for file in files:
        try:
            os.remove(f"{temp_dir}/{get_shortname(file)}.m4a")
        except FileNotFoundError:
            pass
    os.rmdir(temp_dir)

    print("YAY")


# # # GUI Widgets
# # Option Menu
album_image_file = Image.open("empty.png").resize((300, 300))
album_image = ImageTk.PhotoImage(album_image_file)
album_image_label = ttk.Label(root, image=album_image, padding=0)
album_image_label.place(x=0, y=0)

go_button = Button(root, text="R E N D E R", command=startRender, bg="#ff4d4d", fg="#fff", width=12,
                   font=("Arial", "50"))
go_button.place(x=400, y=500)

folder_select_button = Button(root, text="CHOOSE FOLDER", command=getFolder, bg="#363636", fg="#fff",
                              font=("Arial", "30"))
folder_select_button.place(x=450, y=400)

root.mainloop()
