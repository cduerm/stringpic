This is a tool to creat string-art in the style of [Petros Vellos](https://artof01.com/vrellis/works/knit.html) and [knitter](https://github.com/christiansiegel/knitter) written in pure go without any dependencies outside the standard library (except the UI tool using the awesome [fyne](fyne.io) framework).

The tool is still work in proggress, use at own peril.

```
Usage of stringpic:
  -darkness int
        string darkness (value between 1 and 255) (default 32)
  -diameter float
        diameter of ring (for string length calculation) in mm (default 0.226)
  -filename string
        png file to convert to string art
  -nLines int
        number of lines (default 2000)
  -output string
        directory where to put the output files (default "output")
  -pinCount int
        number of pins in circular pattern (default 300)
  -size int
        size of output image (default 512)
```

This project is licensed under the terms of the MIT license (see [LICENSE.md](LICENSE.md))