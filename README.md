# Topography
This is a topographical viewer that renders views of the earth

## Usage
**To use this you will need a dataset in a similair format to SRTM15+, instructions for installing can be found below. Ensure you also have libgdal-dev installed on your system.**

There are two main modes you can run, one of which is a webserver, the other is a renderer. Both require the -f option to be passed for the dataset. And they look like as follows

```sh
topography -f srtm15plus_f16.tif --render
```

```sh
topography -f srtm15plus_f16.tif --server
```

## Dataset
This tool was made for the SRTM15+ dataset, though it should technically work with any GDAL compatible dataset that uses the same pixel layout

To download the SRTM15+ dataset I recommend getting it from [here](https://portal.opentopography.org/raster?opentopoID=OTSRTM.122019.4326.1), once installed I recommend using the script *scripts/data.py* which performs compressions on the dataset and unifies it into a single dataset. I recommend using the --f16 compression, as it'll lead to significantly better performance and is only off by about a meter.