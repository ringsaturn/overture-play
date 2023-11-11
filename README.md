# Have fun with Overture data.

- [Have fun with Overture data.](#have-fun-with-overture-data)
  - [Prepare](#prepare)
  - [POI](#poi)
    - [Setup a nearby search server](#setup-a-nearby-search-server)
  - [Admin](#admin)
    - [Setup a adcode query server](#setup-a-adcode-query-server)
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
mkdir themes
cd themes
mkdir places
cd places
# Download POI data, about 8.6 GB
aws s3 cp --recursive --region us-west-2 --no-sign-request s3://overturemaps-us-west-2/release/2023-10-19-alpha.0/theme=places/ .
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

Build binary:

```bash
cd poi-server;go build;cd ..
```

Run:

```bash
# Run for sample
./poi-server/poi-server -places-file ./places.sample.geojson

# Run for Beijing
./poi-server/poi-server -places-file ./places.beijing.geojson

# Run for all, I don't have enough CPU&memory to run this
./poi-server/poi-server -places-file ./themes/places/places.geojson
```

Call API:

```bash
curl "http://localhost:8888/nearby?lng=116.3529&lat=40.0008&count=50"
```

Output:

- Data: https://gist.github.com/ringsaturn/9f5de41a25937e4d7befb0eab907b76c
- Add to maps:
  <http://geojson.io/#id=gist:ringsaturn/9f5de41a25937e4d7befb0eab907b76c>

## Admin

```bash
mkdir themes
cd themes
mkdir admins
cd admins
# Download admin data, about 600MB
aws s3 cp --recursive --region us-west-2 --no-sign-request s3://overturemaps-us-west-2/release/2023-10-19-alpha.0/theme=admins/ .
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

### Setup a adcode query server

```bash
cd admin-server;go build;cd ..
./admin-server/admin-server -admin-file themes/admins/countries.geojson
```

```bash
curl http://localhost:8888/admin?lng=-77.0391101&lat=38.8976763
```

Output:

```json
{
  "data": [
    {
      "adminlevel": 4,
      "isocountrycodealpha2": null,
      "localitytype": "state",
      "names": {
        "common": [
          { "language": "local", "value": "District of Columbia" },
          { "language": "en-Latn", "value": "District of Columbia" }
        ]
      },
      "sources": [{ "dataset": "TomTom", "property": "" }],
      "subtype": "administrativeLocality",
      "type": "locality"
    },
    {
      "adminlevel": 2,
      "isocountrycodealpha2": "US",
      "localitytype": "country",
      "names": {
        "common": [
          { "language": "local", "value": "United States" },
          { "language": "en-Latn", "value": "United States" },
          { "language": "en", "value": "United States" },
          { "language": "nl", "value": "United States" },
          { "language": "es", "value": "Estados Unidos" },
          { "language": "pt", "value": "Estados Unidos" },
          { "language": "pt-BR", "value": "Estados Unidos" },
          { "language": "fr", "value": "États-Unis" },
          { "language": "de", "value": "Vereinigte Staaten von Amerika" },
          { "language": "ru", "value": "США" },
          { "language": "it", "value": "Stati Uniti" },
          { "language": "ar", "value": "الولايات المتحدة" },
          { "language": "bg", "value": "САЩ" },
          { "language": "bs", "value": "SAD" },
          { "language": "hr", "value": "SAD" },
          { "language": "ca", "value": "Estats Units" },
          { "language": "cs", "value": "Spojené státy" },
          { "language": "da", "value": "USA" },
          { "language": "no", "value": "USA" },
          { "language": "sv", "value": "USA" },
          { "language": "el", "value": "Ηνωμένες Πολιτείες" },
          { "language": "et", "value": "Ameerika Ühendriigid" },
          { "language": "fi", "value": "Yhdysvallat" },
          { "language": "he", "value": "ארצות הברית" },
          { "language": "hu", "value": "Egyesült Államok" },
          { "language": "id", "value": "Amerika Serikat" },
          { "language": "ja", "value": "アメリカ合衆国" },
          { "language": "ko", "value": "미국" },
          { "language": "lt", "value": "JAV" },
          { "language": "lv", "value": "Amerikas Savienotās valstis" },
          { "language": "pl", "value": "Stany Zjednoczone" },
          { "language": "pt-PT", "value": "Estados Unidos da América" },
          { "language": "ro", "value": "Statele Unite" },
          { "language": "sk", "value": "\"Spojené štáty, USA\"" },
          { "language": "sl", "value": "Združene države Amerike" },
          { "language": "sr-Latn", "value": "Sjedinjene Američke Države" },
          { "language": "th", "value": "สหรัฐอเมริกา" },
          { "language": "tr", "value": "Amerika Birleşik Devletleri" },
          { "language": "uk", "value": "Сполучені Штати Америки" },
          { "language": "vi", "value": "Mỹ" },
          { "language": "zh-Hans", "value": "美国" },
          { "language": "zh-Hant", "value": "美國" }
        ]
      },
      "sources": [
        { "dataset": "TomTom", "property": "" },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/2"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/3"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/4"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/5"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/6"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/7"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/8"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/9"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/10"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/11"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/12"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/13"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/14"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/15"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/16"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/17"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/18"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/19"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/20"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/21"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/22"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/23"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/24"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/25"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/26"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/27"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/28"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/29"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/30"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/31"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/32"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/33"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/34"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/35"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/36"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/37"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/38"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/39"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/40"
        },
        {
          "dataset": "Esri Community Maps",
          "property": "/properties/names/common/41"
        }
      ],
      "subtype": "administrativeLocality",
      "type": "locality"
    }
  ]
}
```

## References

- <https://github.com/OvertureMaps/data>
- <https://github.com/bertt/overture>
