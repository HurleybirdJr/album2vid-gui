#!/usr/bin/env python3

# AUTHOR: Nicholas Preston (npgy)
# CO-AUTHOR (GUI): Will Hurley (HurleybirdJr)

from glob import glob
import os
import subprocess
import mutagen
import time
from sys import exit
import shutil
from pathlib import Path
from tkinter import *
from tkinter import filedialog, ttk
from PIL import Image, ImageTk

# TODO: Implement "fast_mode" as GUI option before folder import
fast_mode = True

GLOBAL_DIR = None
GLOBAL_TEMP_PATH = None
RENDER_ARGS = {}
files = []
ffmpeg_binary = "ffmpeg"

# ## Commented assuming @npgy:main PullRequest#2 is valid ##
# Fix command for linux systems
# if platform == "linux":
#     ffmpeg = "./" + ffmpeg



def get_runtime(filename: str) -> float:
    """Tries mutagen first; falls back to ffprobe/ffmpeg if metadata fails."""
    file_path = Path(filename)

    # --- 1. Try Mutagen (Fastest, most reliable for simple tags) ---
    try:
        audio_file = mutagen.File(str(file_path))
        if not audio_file:
            return 0.0

        # Attempt to pull duration from common keys/structures
        duration_keys = ['duration', 'length']
        for key in duration_keys:
            try:
                # Try reading directly from tags or the generic container structure
                duration = audio_file.get(key)
                if duration is None and isinstance(audio_file, dict):
                    duration = audio_file['tags'].get(key) # Fallback to 'tags' dictionary if needed

                if duration:
                    return float(duration)
            except Exception as e:
                # If this specific key fails, continue trying others
                pass

    except mutagen.MutagenError as e:
        print(f"Warning: Mutagen failed for {file_path.name}. Error: {e}")
    except Exception as e:
        print(f"Warning: General error during metadata check for {file_path.name}. Error: {e}")

    # --- 2. Fallback to ffprobe (Most accurate, recommended) ---
    # We use ffprobe if it exists and is robustly called.
    try:
        # Check if ffprobe command exists first
        subprocess.run(['ffprobe', '-version'], capture_output=True, check=True)

        print(f"Attempting detailed runtime check using ffprobe for {file_path.name}...")

        # Use json output format and ask only for the duration
        cmd = [
            'ffprobe',
            '-v', 'error',
            '-show_entries', 'format=duration', # We want the container duration
            '-of', 'json',  # Output in JSON format for easy parsing
            str(file_path)
        ]

        result = subprocess.run(cmd, capture_output=True, text=True, check=True)
        import json
        data = json.loads(result.stdout)

        duration_str = data['format'].get('duration')
        if duration_str:
            return float(duration_str)

    except FileNotFoundError:
        # ffprobe was not found, proceed to basic ffmpeg fallback
        pass
    except subprocess.CalledProcessError as e:
        print(f"Warning: ffprobe failed for {file_path.name}. Error: {e}")
    except json.JSONDecodeError:
        print("Warning: Failed to parse JSON output from ffprobe.")


    # --- 3. Fallback to basic ffmpeg (If neither of the above works) ---
    try:
        print(f"Attempting final fallback check using ffmpeg for {file_path.name}...")
        cmd = [
            'ffmpeg',
            '-v', 'error',
            # This command simply reads headers and prints stream info to stderr,
            # which is usually where duration lives.
            '-i', str(file_path)
        ]

        subprocess.run(cmd, capture_output=True, check=True)
        # Note: Parsing the raw output of ffmpeg is highly complex and unreliable.
        # If we reach this point, it's safer to just assume failure.
    except Exception as e:
        print(f"Warning: All metadata checks failed for {file_path.name}. Error: {e}")

    return 0.0 # Return 0 if all attempts fail


def get_shortname(filename: Path) -> str:
    """Returns just the file's name

    ;param filename: the file's full path
    ;return: the file's name
    """
    return filename.stem


def get_timestamp(seconds):
    """Returns the formatted timestamp

    ;param seconds: the length of time in seconds
    ;return: the formatted timestamp
    """
    return time.strftime("%H:%M:%S", time.gmtime(seconds))


def throw_error(text):
    print("ERROR: " + text)
    exit(1)


def cleanup():
    """
    Cleans up temporary files that may have been generated
    """
    global GLOBAL_TEMP_PATH
    # Try deleting temp directory
    if GLOBAL_TEMP_PATH and GLOBAL_TEMP_PATH.exists():
        try:
            shutil.rmtree(GLOBAL_TEMP_PATH, ignore_errors=True, onerror=None)
            print("Cleaned up temp directory.")
        except Exception as e:
            print(f"Warning during cleanup: {e}")

    manifest_path = Path(GLOBAL_DIR) / "files.txt"
    if manifest_path.exists():
        os.remove(manifest_path)
        print("Removed file list manifest.")

# # # GUI Functions
def getFolder():
    """Handles directory selection and file preparation."""
    global GLOBAL_DIR, GLOBAL_TEMP_PATH, RENDER_ARGS

    selected_dir = filedialog.askdirectory()
    if not selected_dir:
        return None

    GLOBAL_DIR = Path(selected_dir)
    GLOBAL_TEMP_PATH = GLOBAL_DIR / ".temp"

    # Find all audio files in the directory
    audio_extensions = ('wav', 'mp3', 'm4a', 'ogg', 'flac')
    all_files = []
    for ext in audio_extensions:
        all_files.extend(list(GLOBAL_DIR.glob(f"*.{ext}")))
    if not all_files:
        throw_error("No supported audio files could be found.")

    global RENDER_FILES
    RENDER_FILES = sorted(all_files)
    print(f"Found {len(RENDER_FILES)} tracks.")

    # Find cover art
    cover_paths = list(GLOBAL_DIR.glob("cover.jpg")) + list(GLOBAL_DIR.glob("cover.png"))
    if not cover_paths:
        throw_error("The cover .png/.jpg photo could not be found.")

    cover_path = Path(cover_paths[0])

    album2_image_file = ImageTk.PhotoImage(Image.open(cover_path).resize((300, 300)))
    album_image_label.configure(image=album2_image_file)
    album_image_label.image = album2_image_file

    # Clean up temp files in case program was exited abruptly on last run
    cleanup()

    # Calculate tracklist with timestamps
    total_runtime = 0.0
    manifest_path = GLOBAL_DIR / "tracklist.txt"
    with open(manifest_path, "w") as f:
        for file_path in RENDER_FILES:
            duration = get_runtime(file_path)
            shortname = file_path.stem
            f.write(f"{shortname} -- {get_timestamp(total_runtime)}\n")
            total_runtime += duration

    # Write all audio files to a temporary text document for ffmpeg
    manifest_file = GLOBAL_DIR / "files.txt"
    with open(manifest_file, "w") as f:
        for file_path in RENDER_FILES:
            f.write(f"file '{str(file_path)}'\n")

    go_button.configure(bg="#a9ff4d")

    RENDER_ARGS = {
        "cover": str(cover_path),
        "manifest_file": str(manifest_file),
        "total_duration": total_runtime
    }

    return True

def startRender():
    """Executes FFmpeg using a robust argument list."""

    global RENDER_ARGS, GLOBAL_DIR

    if not RENDER_ARGS:
        throw_error("Please select a folder first before attempting to render.")

    total_duration = RENDER_ARGS["total_duration"]
    cover_path = RENDER_ARGS["cover"]
    manifest_file = RENDER_ARGS["manifest_file"]
    output_path = GLOBAL_DIR / "out.mp4"

    if total_duration <= 0:
        throw_error("Cannot render video because the calculated total track duration is zero or invalid.")

    # Gonna comment since even I struggle to grasp ffmpeg arguments lol
    cmd_args = [
        ffmpeg_binary,
        "-hwaccel", "auto", "-y",

        # Input 1: Cover Image
        "-loop", "1", "-framerate", "1", "-i", cover_path,

        # Input 2: Audio Manifest List
        "-f", "concat", "-safe", "0", "-i", str(manifest_file),
    ]

    duration_str = f"{total_duration:.6f}"
    filtergraph = (
        f"[0:v]trim=duration={duration_str},format=yuv420p,scale=1080:1080[v]"
    )
    cmd_args.extend([
        "-filter_complex", filtergraph,
        "-map", "[v]",
        "-map", "1:a",
        "-b:a", "320k",
        "-shortest",
        "-tune", "stillimage",
        str(output_path)
    ])

    go_button.configure(bg="#a9ff4d")

    # Execute FFMPEG command and delete temporary file
    try:
        subprocess.run(cmd_args, check=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
        print("YAY")
    except subprocess.CalledProcessError as e:
        full_error = f"FFmpeg failed (Exit Code {e.returncode}).\n"
        full_error += "--- STDOUT ---\n" + e.stdout.decode("utf-8")
        throw_error(f"Detailed FFmpeg error:\n{full_error}")
    except Exception as e:
        throw_error(f"An unexpected error occurred while running ffmpeg: {e}")

    # Clean up temp files
    cleanup()

def setup_gui():
    # # # GUI # # #
    root = Tk()

    # # # GUI attributes
    root.title("album2vid-gui")
    root.resizable(False, False)
    root["bg"] = "#1c1c1c"

    root.update_idletasks()

    gui_width = 1000
    gui_height = 700
    # Locate display center from display dimensions
    center_x = int(root.winfo_screenwidth() / 2 - gui_width / 2)
    center_y = int(root.winfo_screenheight() / 2 - gui_height / 2)
    # Position GUI to display center
    root.geometry(f'{gui_width}x{gui_height}+{center_x}+{center_y}')

    # # # GUI Widgets



    # # Option Menu
    folder_select_button = Button(root, text="CHOOSE FOLDER", command=getFolder, bg="#363636", fg="#fff",
                                  font=("Arial", "30"))
    folder_select_button.place(x=450, y=400)

    global go_button
    go_button = Button(root, text="R E N D E R", command=startRender, bg="#ff4d4d", fg="#fff", width=12,
                       font=("Arial", "50"))
    go_button.place(x=400, y=500)



    return root

if __name__ == "__main__":
    print("Welcome to album2vid!")

    root = setup_gui()

    # # Image Art Preview
    album_image_file = Image.open("empty.png").resize((300, 300))
    album_image = ImageTk.PhotoImage(album_image_file)
    global album_image_label
    album_image_label = ttk.Label(root, image=album_image, padding=0)
    album_image_label.place(x=0, y=0)

    root.mainloop()
