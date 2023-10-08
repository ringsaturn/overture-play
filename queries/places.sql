-- Install via `INSTALL spatial;`
LOAD spatial;
-- Create file for all places.
COPY (
    SELECT id,
        updatetime,
        version,
        CAST(names AS JSON) AS names,
        CAST(categories AS JSON) AS categories,
        confidence,
        CAST(websites AS JSON) AS websites,
        CAST(socials AS JSON) AS socials,
        CAST(emails AS JSON) AS emails,
        CAST(phones AS JSON) AS phones,
        CAST(brand AS JSON) AS brand,
        CAST(addresses AS JSON) AS addresses,
        CAST(sources AS JSON) AS sources,
        ST_GeomFromWKB(geometry)
    from read_parquet(
            'type=place/*',
            filename = true,
            hive_partitioning = 1
        )
) TO 'places.geojson' WITH (FORMAT GDAL, DRIVER 'GeoJSON');
-- Create file for part of Beijing
COPY (
    SELECT id,
        updatetime,
        version,
        CAST(names AS JSON) AS names,
        CAST(categories AS JSON) AS categories,
        confidence,
        CAST(websites AS JSON) AS websites,
        CAST(socials AS JSON) AS socials,
        CAST(emails AS JSON) AS emails,
        CAST(phones AS JSON) AS phones,
        CAST(brand AS JSON) AS brand,
        CAST(addresses AS JSON) AS addresses,
        CAST(sources AS JSON) AS sources,
        ST_GeomFromWKB(geometry)
    from read_parquet(
            'type=place/*',
            filename = true,
            hive_partitioning = 1
        )
    WHERE bbox.minX > 116.2908
        and bbox.maxX < 116.5263
        and bbox.minY > 39.8555
        and bbox.maxY < 40.0219
) TO 'places.beijing.geojson' WITH (FORMAT GDAL, DRIVER 'GeoJSON');
-- Create file for testing
COPY (
    SELECT id,
        updatetime,
        version,
        CAST(names AS JSON) AS names,
        CAST(categories AS JSON) AS categories,
        confidence,
        CAST(websites AS JSON) AS websites,
        CAST(socials AS JSON) AS socials,
        CAST(emails AS JSON) AS emails,
        CAST(phones AS JSON) AS phones,
        CAST(brand AS JSON) AS brand,
        CAST(addresses AS JSON) AS addresses,
        CAST(sources AS JSON) AS sources,
        ST_GeomFromWKB(geometry)
    from read_parquet(
            'type=place/*',
            filename = true,
            hive_partitioning = 1
        )
    WHERE bbox.minX > 5.1
        and bbox.maxX < 5.2
        and bbox.minY > 52.1
        and bbox.maxY < 52.2
) TO 'places.geojson' WITH (FORMAT GDAL, DRIVER 'GeoJSON');