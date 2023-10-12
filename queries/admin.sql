LOAD spatial;
--- Process all
COPY (
    SELECT type,
        subType,
        localityType,
        adminLevel,
        isoCountryCodeAlpha2,
        JSON(names) AS names,
        JSON(sources) AS sources,
        ST_GeomFromWkb(geometry) AS geometry
    FROM read_parquet(
            './type=*/*',
            filename = true,
            hive_partitioning = 1
        )
    WHERE ST_GeometryType(ST_GeomFromWkb(geometry)) IN ('POLYGON', 'MULTIPOLYGON')
) TO 'countries.geojson' WITH (FORMAT GDAL, DRIVER 'GeoJSON');
-- Process level 2
COPY (
    SELECT type,
        subType,
        localityType,
        adminLevel,
        isoCountryCodeAlpha2,
        JSON(names) AS names,
        JSON(sources) AS sources,
        ST_GeomFromWkb(geometry) AS geometry
    FROM read_parquet(
            './type=*/*',
            filename = true,
            hive_partitioning = 1
        )
    WHERE adminLevel = 2
        AND ST_GeometryType(ST_GeomFromWkb(geometry)) IN ('POLYGON', 'MULTIPOLYGON')
) TO 'countries_level_2.geojson' WITH (FORMAT GDAL, DRIVER 'GeoJSON');