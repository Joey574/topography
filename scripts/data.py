import numpy as np
import os
from osgeo import gdal

gdal.UseExceptions()
gdal.SetConfigOption('GDAL_NUM_THREADS', 'ALL_CPUS')
gdal.SetConfigOption('GDAL_CACHEMAX', '512')

def create_dataset(original_path, converted_path, use_f16=True):
    ds = gdal.Open(original_path)

    options = gdal.TranslateOptions()
    if use_f16:
        options = gdal.TranslateOptions(
            format="GTiff",
            outputType=gdal.GDT_Float16,
            creationOptions=[
                "COMPRESS=ZSTD",
                "ZSTD_LEVEL=9",
                "NUM_THREADS=ALL_CPUS",
                "DISCARD_LSB=1",
                "TILED=YES",
                "BLOCKXSIZE=1024",
                "BLOCKYSIZE=1024",
            ]
        )
    else:
        options = gdal.TranslateOptions(
            format="GTiff",
            outputType=gdal.GDT_Float32,
            creationOptions=[
                "COMPRESS=LERC_ZSTD",
                "MAX_Z_ERROR=1.0",
                "DISCARD_LSB=2",
                "NUM_THREADS=ALL_CPUS",            
                "TILED=YES",
                "BLOCKXSIZE=1024",
                "BLOCKYSIZE=1024",
            ]
        )

    gdal.Translate(converted_path, ds, options=options)

    print(f"--- Dataset Generated ---")
    print(f"New Size: {os.path.getsize(converted_path) / (1024**3)} gb")
    

def audit_dataset(original_path, converted_path, block_size=4096):
    # Open both datasets
    ds_orig = gdal.Open(original_path)
    ds_conv = gdal.Open(converted_path)
    
    x_size = ds_orig.RasterXSize
    y_size = ds_orig.RasterYSize
    
    print()
    print(f"Dataset Size: {x_size} x {y_size}")
    print(f"Block Size: {block_size} x {block_size}")
    
    # Initialize Statistics
    # Using infinity ensures the first real value encountered becomes the new min/max
    org_range = [float('inf'), float('-inf')]
    mod_range = [float('inf'), float('-inf')]
    
    sum_sq_diff = 0.0
    sum_diff = 0.0
    max_err = 0.0
    total_pixels = 0

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
    
    # Final Calculations
    rmse = np.sqrt(sum_sq_diff / total_pixels)
    mean_bias = sum_diff / total_pixels

    print(f"--- Precision Audit Results ---")
    print(f"Original Range: {org_range}")
    print(f"Modified Range: {mod_range}")
    print(f"RMSE: {rmse:.6f} units")
    print(f"Max Absolute Error: {max_err:.6f} units")
    print(f"Mean Bias: {mean_bias:.6e} units")
    print(f"Total Pixels Audited: {total_pixels}")

original = "/home/joey574/Downloads/SRTM15Plus/SRTM15Plus_srtm.vrt"

modified_f16 = "/home/joey574/repos/topography/datasets/srtm15plus_f16.tif"
modified_f32 = "/home/joey574/repos/topography/datasets/srtm15plus_f32.tif"

# create_dataset(original, modified_f16, True)
audit_dataset(original, modified_f16)

# create_dataset(original, modified_f32, False)
# audit_dataset(original, modified_f32)
