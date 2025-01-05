# scraperlite

Scrape text and HTML based on CSS selectors and save contents to a SQLite database.

Repeated runs save changed content and the observation timestamp.

## Example

``` shell
scraperlite https://go.dev \
    popularCLIPackages.html '#main-content > section.WhyGo > div > ul > li:nth-child(2) > div.WhyGo-reasonFooter > div.WhyGo-reasonPackages > ul' \
    whyWebDevelopment.txt '#main-content > section.WhyGo > div > ul > li:nth-child(3) > div.WhyGo-reasonDetails > div.WhyGo-reasonText > p'
```

In a sqlite3 shell:

``` shell
sqlite> select t, substr(json_extract(content, '$.popularCLIPackages.html'), 1, 20) || '...' as popular_packages_html,
  json_extract(content, '$.whyWebDevelopment.txt') as why_web_development
  from observations join contents on (contents.id=content_id)
  order by t;
+----------------------------------+-------------------------+-----------------------------------------------------------+
|                t                 |  popular_packages_html  |                    why_web_development                    |
+----------------------------------+-------------------------+-----------------------------------------------------------+
| 2025-01-05T18:59:27.496327-04:00 | <div class="WhyGo-re... | With enhanced memory performance and support for several  |
|                                  |                         | IDEs, Go powers fast and scalable web applications.       |
+----------------------------------+-------------------------+-----------------------------------------------------------+
```
