LOAD spatial;
-- Full
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
) TO 'countries_all_levels.geojson' WITH (FORMAT GDAL, DRIVER 'GeoJSON');
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
) TO 'countries.geojson' WITH (FORMAT GDAL, DRIVER 'GeoJSON');
-- Process level 4
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
    WHERE adminLevel = 4
        AND ST_GeometryType(ST_GeomFromWkb(geometry)) IN ('POLYGON', 'MULTIPOLYGON')
) TO 'countries_level4.geojson' WITH (FORMAT GDAL, DRIVER 'GeoJSON');