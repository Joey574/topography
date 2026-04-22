import numpy as np
import os
from osgeo import gdal
import argparse

gdal.UseExceptions()
gdal.SetConfigOption('GDAL_NUM_THREADS', 'ALL_CPUS')

def create_dataset(original_path, converted_path, use_f16, downsample, dont_compress):
    ds = gdal.Open(original_path)
    if not ds:
        raise ValueError("dataset not found")

    kwargs = {
        "format":"GTiff"
    }

    if use_f16:
        kwargs["outputType"] = gdal.GDT_Float16

        if not dont_compress:
            kwargs["creationOptions"] = [
                "COMPRESS=ZSTD", "ZSTD_LEVEL=9", "NUM_THREADS=ALL_CPUS",
                "DISCARD_LSB=2", "TILED=YES", "BLOCKXSIZE=2048", "BLOCKYSIZE=2048"
            ]
    else:
        kwargs["outputType"] = gdal.GDT_Float32

        if not dont_compress:
            kwargs["creationOptions"] = [
                "COMPRESS=LERC_ZSTD", "MAX_Z_ERROR=1.0", "DISCARD_LSB=2",
                "NUM_THREADS=ALL_CPUS", "TILED=YES", "BLOCKXSIZE=2048", "BLOCKYSIZE=2048"
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

parser = argparse.ArgumentParser(description="Simple tool to support dataset compression and auditing for Joey574/topography")
parser.add_argument("-f", "--file", type=str, action='append')
parser.add_argument("-o", "--output", type=str)
parser.add_argument("--f16", action="store_true")
parser.add_argument("--f32", action="store_true")
parser.add_argument("--audit", action="store_true")
parser.add_argument("-d", "--downsample", type=int)
parser.add_argument("--no-compression", action="store_true")
args = parser.parse_args()

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