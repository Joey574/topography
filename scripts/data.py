#!/bin/python3

import numpy as np
import os
from osgeo import gdal, osr
import argparse

gdal.UseExceptions()
gdal.SetConfigOption('GDAL_NUM_THREADS', 'ALL_CPUS')

def get_lat_long_extent(ds):
    """Calculates the Min/Max Lat and Long for a dataset."""
    gt = ds.GetGeoTransform()
    width = ds.RasterXSize
    height = ds.RasterYSize

    # 1. Get coordinates of the 4 corners in the dataset's native CRS
    # Coordinates: (Long-ish/X, Lat-ish/Y)
    # Extent: [Left, Top, Right, Bottom]
    ext = [
        gt[0],                     # Top Left X
        gt[3],                     # Top Left Y
        gt[0] + width * gt[1],     # Bottom Right X
        gt[3] + height * gt[5]     # Bottom Right Y
    ]

    # 2. Setup Coordinate Transformation to WGS84 (Lat/Long)
    src_srs = osr.SpatialReference()
    src_srs.ImportFromWkt(ds.GetProjection())
    
    tgt_srs = osr.SpatialReference()
    tgt_srs.ImportFromEPSG(4326) # WGS84
    
    # Ensure Axis Mapping is correct for modern GDAL (Long, Lat vs Lat, Long)
    tgt_srs.SetAxisMappingStrategy(osr.OAMS_TRADITIONAL_GIS_ORDER)
    
    transform = osr.CoordinateTransformation(src_srs, tgt_srs)

    # 3. Transform corners to Lat/Long
    # (min_x, max_y) -> Top Left
    # (max_x, min_y) -> Bottom Right
    ul = transform.TransformPoint(ext[0], ext[1])
    lr = transform.TransformPoint(ext[2], ext[3])

    return {
        "min_lat": min(ul[0], lr[0]),
        "max_lat": max(ul[0], lr[0]),
        "min_lon": min(ul[1], lr[1]),
        "max_lon": max(ul[1], lr[1])
    }

def create_dataset(original_path, converted_path, use_f16, downsample, dont_compress):
    ds = gdal.Open(original_path)
    if not ds:
        raise ValueError("dataset not found")

    kwargs = {
        "format":"GTiff",
        "creationOptions":["BIGTIFF=YES"]
    }

    if use_f16:
        kwargs["outputType"] = gdal.GDT_Float16

        if not dont_compress: # compress
            kwargs["creationOptions"] = [
                "COMPRESS=ZSTD", "ZSTD_LEVEL=9", "NUM_THREADS=ALL_CPUS",
                "DISCARD_LSB=1", "TILED=YES", "BLOCKXSIZE=2048", "BLOCKYSIZE=2048",
                "BIGTIFF=YES"
            ]
    else:
        kwargs["outputType"] = gdal.GDT_Float32

        if not dont_compress: # compress
            kwargs["creationOptions"] = [
                "COMPRESS=LERC_ZSTD", "MAX_Z_ERROR=1.0", "DISCARD_LSB=1",
                "NUM_THREADS=ALL_CPUS", "TILED=YES", "BLOCKXSIZE=2048", "BLOCKYSIZE=2048",
                "BIGTIFF=YES"
            ]

    if downsample:
        orig_width = ds.RasterXSize
        orig_height = ds.RasterYSize
        target_height = int(round(orig_height * (downsample / orig_width)))

        kwargs["width"] = downsample
        kwargs["height"] = target_height

        kwargs["resampleAlg"] = gdal.GRA_Average

    print("Creating dataset...")
    options = gdal.TranslateOptions(**  kwargs)
    gdal.Translate(converted_path, ds, options=options)
    ds = None

    print("--- Dataset Generated ---")
    print(f"New Size: {os.path.getsize(converted_path) / (1024**3)} gb")

def audit_dataset(original_path, converted_path, block_size=4096):
    # Open both datasets
    ds_orig = gdal.Open(original_path)
    ds_conv = gdal.Open(converted_path)

    if ds_orig == None or ds_conv == None:
        print("Failed to open datasets")
        return
    
    x_size = ds_orig.RasterXSize
    y_size = ds_orig.RasterYSize
    
    org_range = [float('inf'), float('-inf')]
    mod_range = [float('inf'), float('-inf')]
    
    sum_sq_diff = 0.0
    sum_diff = 0.0
    max_err = 0.0
    total_pixels = 0

    print("Auditing datasets...")

    # Iterate through the entire grid
    for y in range(0, y_size, block_size):
        for x in range(0, x_size, block_size):
            
            # Adjust window size for edge blocks
            win_x = min(block_size, x_size - x)
            win_y = min(block_size, y_size - y)
            
            # Read data and cast to float64 to prevent overflow/precision loss during math
            orig_data = ds_orig.GetRasterBand(1).ReadAsArray(x, y, win_x, win_y).astype(np.float64)
            conv_data = ds_conv.GetRasterBand(1).ReadAsArray(x, y, win_x, win_y).astype(np.float64)
            
            # Calculate Difference
            diff = orig_data - conv_data

            # Update Ranges
            org_range[0] = min(org_range[0], np.min(orig_data))
            org_range[1] = max(org_range[1], np.max(orig_data))
            
            mod_range[0] = min(mod_range[0], np.min(conv_data))
            mod_range[1] = max(mod_range[1], np.max(conv_data))
            
            # Update Statistics Accumulators
            sum_sq_diff += np.sum(diff**2)
            sum_diff += np.sum(diff)
            max_err = max(max_err, np.max(np.abs(diff)))
            total_pixels += diff.size

        # Simple progress indicator for rows
        progress = (y + win_y) / y_size * 100
        print(f"\rProgress: {progress:.1f}%", end='')

    ds_orig.Close()
    ds_conv.Close()
    print("\n" + "-"*25)
    
    # Finalize statistics
    rmse = np.sqrt(sum_sq_diff / total_pixels)
    mean_bias = sum_diff / total_pixels

    print(f"--- Precision Audit Results ---")
    print(f"Original Range: {org_range}")
    print(f"Modified Range: {mod_range}")
    print(f"RMSE: {rmse:.6f} meters")
    print(f"Max Absolute Error: {max_err:.6f} meters")
    print(f"Mean Bias: {mean_bias:.6e} meters")

def dataset_info(files):
    for f in files:
        try:
            ds = gdal.Open(f)
            if ds is None:
                print(f"Could not open: {f}")
                continue


            print(f"\n{'='*40}")
            print(f"PATH: {f}")
            print(f"{'='*40}")

            driver = ds.GetDriver()
            if driver:
                print(f"Driver:       {driver.ShortName} ({driver.LongName})")

            print(f"Size (X, Y):  {ds.RasterXSize} x {ds.RasterYSize}")
            print(f"Bands:        {ds.RasterCount}")

            # Get Extent
            try:
                extent = get_lat_long_extent(ds)
                print(f"Longitude Range: {extent['min_lon']:.6f} to {extent['max_lon']:.6f}")
                print(f"Latitude Range:  {extent['min_lat']:.6f} to {extent['max_lat']:.6f}")
            except Exception as e:
                pass

            data = ds.GetRasterBand(1).ReadAsArray(0, 0, ds.RasterXSize, ds.RasterYSize)
            minv = np.min(data)
            maxv = np.max(data)
            print(f"Value Range: [{minv:.6f}, {maxv:.6f}]")

            if ds.RasterCount > 0:
                band = ds.GetRasterBand(1)
                dtype = gdal.GetDataTypeName(band.DataType)
                print(f"Data Type:    {dtype}")

            gt = ds.GetGeoTransform()
            if gt:
                print(f"Origin:       ({gt[0]:.2f}, {gt[3]:.2f})")
                print(f"Pixel Size:   ({gt[1]:.2f}, {gt[5]:.2f})")

            proj = ds.GetProjection()
            if proj:
                print(f"Projection:   {proj[:60]}...")

        except Exception as e:
            print(f"Error processing {f}: {e}")
        
        finally:
            ds = None

parser = argparse.ArgumentParser(description="Simple tool to support dataset compression and auditing for github.com/Joey574/topography")
parser.add_argument("-f", "--file", type=str, action='append')
parser.add_argument("-o", "--output", type=str)
parser.add_argument("--f16", action="store_true")
parser.add_argument("--f32", action="store_true")
parser.add_argument("--audit", action="store_true")
parser.add_argument("--info", action="store_true")
parser.add_argument("--no-auto-detect", action="store_true", help="Disable auto-detection for --patch")
parser.add_argument("-d", "--downsample", type=int)
parser.add_argument("--no-compression", action="store_true")
args = parser.parse_args()

if args.info:
    if args.file == None:
        print("pass the paths of the datasets with -f dataset1")
        exit(1)
    dataset_info(args.file)    
    exit(0)

if args.audit:
    if args.file == None or len(args.file) != 2:
        print("pass the paths to the original and mofied datasets with -f path/to/dataset1 -f path/to/dataset2")
        exit(1)
    audit_dataset(args.file[0], args.file[1])
    exit(0)

if args.file == None or len(args.file) != 1:
    print("pass the original dataset with -f path/to/dataset")
    exit(1)

if args.output == None:
    print("pass the path for the new dataset with -o some_path")
    exit(1)

if (args.f16 == False and args.f32 == False) or ():
    print("MUST pass --f16 OR --f32 (f16 will result in a smaller size, f32 will result in higher accuracy)")
    exit(1)

create_dataset(args.file[0], args.output, args.f16, args.downsample, args.no_compression)