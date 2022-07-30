# WebM-Dynamic-Resolution
Creates WebMs with dynamic resolutions from videos.
![](https://thumbs.gfycat.com/BriefLikelyIndiancow-size_restricted.gif)

# Usage
**FFmpeg is needed.**  
[Windows (gpl)](https://github.com/BtbN/FFmpeg-Builds/releases)    
Linux: `sudo apt install ffmpeg`    
Termux `pkg install ffmpeg`

`webm_dr_x64.exe in.mkv -o out.webm`

```
Usage: webm_dr_x64.exe [--mode MODE] --outpath OUTPATH INPATH

Positional arguments:
  INPATH

Options:
  --mode MODE, -m MODE   1 = random, 2 = growing. [default: 1]
  --outpath OUTPATH, -o OUTPATH
                         Path to write output WebM to. Path will be made if it doesn't already exist.
  --help, -h             display this help and exit
  ```
