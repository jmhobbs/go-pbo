[![Go Reference](https://pkg.go.dev/badge/github.com/jmhobbs/go-pbo.svg)](https://pkg.go.dev/github.com/jmhobbs/go-pbo)
[![Lint & Test](https://github.com/jmhobbs/go-pbo/actions/workflows/lint-and-test.yml/badge.svg)](https://github.com/jmhobbs/go-pbo/actions/workflows/lint-and-test.yml)
[![codecov](https://codecov.io/github/jmhobbs/go-pbo/graph/badge.svg?token=sB2axgNro5)](https://codecov.io/github/jmhobbs/go-pbo)

# go-pbo

A Go library for working with PBO files from Bohemia Interactive.

## pbotool

Included is a command line tool for manipulating PBO files.

```
COMMAND
  pbotool

USAGE
  pbotool SUBCOMMAND ...

SUBCOMMANDS
  inspect   Inspect the contents of a PBO
  unpack    Unpack all files from a PBO
  pack      Pack files into a PBO
  version   Print version and build date information
```

### Inspect

```
COMMAND
  inspect -- Inspect the contents of a PBO

USAGE
  pbotool inspect <file.pbo>
```

```
$ pbotool inspect ViralSuppresor.pbo
[Properties]
- product: dayz ugc
- prefix: ViralSuppresor

[Files]
- config.cpp (2138 bytes)
- data\Viral_AK_Sup.paa (601349 bytes)
- data\Viral_AK_Sup.png (1612340 bytes)
- scripts\4_World\Classes\recipe\CraftUniSuppressor.c (4131 bytes)
- scripts\4_World\Classes\recipe\PluginRecipesManagerBase.c (198 bytes)
- texHeaders.bin (208 bytes)
```

### Unpack

```
COMMAND
  unpack -- Unpack all files from a PBO

USAGE
  pbotool unpack <file.pbo> <output directory>
```

```
$ pbotool unpack ViralSuppresor.pbo ViralSuppresor
2026/03/11 12:54:53 Unpacking config.cpp
2026/03/11 12:54:53 Unpacking data\Viral_AK_Sup.paa
2026/03/11 12:54:53 Unpacking data\Viral_AK_Sup.png
2026/03/11 12:54:53 Unpacking scripts\4_World\Classes\recipe\CraftUniSuppressor.c
2026/03/11 12:54:53 Unpacking scripts\4_World\Classes\recipe\PluginRecipesManagerBase.c
2026/03/11 12:54:53 Unpacking texHeaders.bin

$ tree ViralSuppresor
ViralSuppresor
├── config.cpp
├── data
│   ├── Viral_AK_Sup.paa
│   └── Viral_AK_Sup.png
├── scripts
│   └── 4_World
│       └── Classes
│           └── recipe
│               ├── CraftUniSuppressor.c
│               └── PluginRecipesManagerBase.c
└── texHeaders.bin

6 directories, 6 files
```

### Pack

```
COMMAND
  pack -- Pack files into a PBO

USAGE
  pbotool pack [FLAGS] <file.pbo> <input file>...

FLAGS
  -p, --property STRING   PBO property in the form of key=value. Can be repeated.
  -r, --recursive         Recursively add files from directories.
```

```
$ pbotool pack -p "product=dayz ugc" -p prefix=ViralSuppresor -r out.pbo ViralSuppresor
+ ViralSuppresor/config.cpp
+ ViralSuppresor/data/Viral_AK_Sup.paa
+ ViralSuppresor/data/Viral_AK_Sup.png
+ ViralSuppresor/scripts/4_World/Classes/recipe/CraftUniSuppressor.c
+ ViralSuppresor/scripts/4_World/Classes/recipe/PluginRecipesManagerBase.c
+ ViralSuppresor/texHeaders.bin
```

# References

https://community.bistudio.com/wiki/PBO_File_Format
