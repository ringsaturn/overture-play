# <h1>Have fun with Overture data in DuckDB.</h1>

- [Have fun with Overture data in DuckDB.](#have-fun-with-overture-data-in-duckdb)
  - [Prepare](#prepare)
  - [POI](#poi)
    - [Count lines](#count-lines)
  - [Admin](#admin)
  - [References](#references)

## Prepare

- Install [`aws-cli`](https://github.com/aws/aws-cli), since we need to download
  data from S3
- Install DuckDB
  - Install DuckDB extension
    - `INSTALL spatial`
  - Open console of DuckDB and load extension:
    - `LOAD 'spatial'`

## POI

```bash
# Download POI data, about 8.6 GB
aws s3 cp --recursive --region us-west-2 --no-sign-request s3://overturemaps-us-west-2/release/2023-07-26-alpha.0/theme=places/ .
```

```bash
# Covert to GeoJSON, about 12 GB
COPY (
SELECT ST_GeomFromWkb(geometry) AS geometry, JSON(names) AS names
from  read_parquet('type=place/*', filename=true, hive_partitioning=1)
) TO 'places.geojson'
WITH (FORMAT GDAL, DRIVER 'GeoJSON');
```

### Count lines

```bash
wc -l places.geojson
```

Output as:

```
59175726
```

## Admin

```bash
# Download admin data, about 600MB
aws s3 cp --recursive --region us-west-2 --no-sign-request s3://overturemaps-us-west-2/release/2023-07-26-alpha.0/theme=admins/ .
```

```sql
COPY (
    SELECT
           type,
           subType,
           localityType,
           adminLevel,
           isoCountryCodeAlpha2,
           JSON(names) AS names,
           JSON(sources) AS sources,
           ST_GeomFromWkb(geometry) AS geometry
      FROM read_parquet('./type=*/*', filename=true, hive_partitioning=1)
     WHERE adminLevel = 2
       AND ST_GeometryType(ST_GeomFromWkb(geometry)) IN ('POLYGON','MULTIPOLYGON')
) TO 'countries.geojson'
WITH (FORMAT GDAL, DRIVER 'GeoJSON');
```

## References

- <https://github.com/OvertureMaps/data>
- <https://github.com/bertt/overture>
