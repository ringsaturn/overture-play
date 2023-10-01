# Have fun with Overture data.

- [Have fun with Overture data.](#have-fun-with-overture-data)
  - [Prepare](#prepare)
  - [POI](#poi)
    - [Setup a nearby search server](#setup-a-nearby-search-server)
  - [Admin](#admin)
  - [References](#references)

## Prepare

- Install [`aws-cli`](https://github.com/aws/aws-cli), since we need to download
  data from S3
- Install DuckDB, we use this to covert original data to GeoJSON if needed
  - Install DuckDB extension
    - `INSTALL spatial`
  - Open console of DuckDB and load extension:
    - `LOAD spatial;`
- Install Go

## POI

```bash
# Download POI data, about 8.6 GB
aws s3 cp --recursive --region us-west-2 --no-sign-request s3://overturemaps-us-west-2/release/2023-07-26-alpha.0/theme=places/ .
```

```sql
COPY (
SELECT 
  id,
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
from  read_parquet('type=place/*', filename=true, hive_partitioning=1)
) TO 'places.geojson'
WITH (FORMAT GDAL, DRIVER 'GeoJSON');
```

Cut a small piece of data for testing:

```sql
COPY (
SELECT 
  id,
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
from  read_parquet('type=place/*', filename=true, hive_partitioning=1)
WHERE bbox.minX > 5.1 and bbox.maxX < 5.2 and bbox.minY>52.1 and bbox.maxY<52.2
) TO 'places.geojson'
WITH (FORMAT GDAL, DRIVER 'GeoJSON');
```

Cut sample data set around Beijing:

```sql
COPY (
SELECT 
  id,
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
from  read_parquet('type=place/*', filename=true, hive_partitioning=1)
WHERE bbox.minX > 116.2908 and bbox.maxX < 116.5263 and bbox.minY>39.8555 and bbox.maxY<40.0219
) TO 'places.beijing.geojson'
WITH (FORMAT GDAL, DRIVER 'GeoJSON');
```

### Setup a nearby search server

```bash
cd poi-server;go build;cd ..

# Run for sample
./poi-server/poi-server -places-file ./places.sample.geojson

# Run for Beijing
./poi-server/poi-server -places-file ./places.beijing.geojson

# Run for all, I don't have enough memory to run this
./poi-server/poi-server -places-file ./places.geojson
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
