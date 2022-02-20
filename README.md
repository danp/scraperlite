# scraperlite

Scrape text and HTML based on CSS selectors and save contents to a SQLite database.

Repeated runs save changed content and the observation timestamp.

## Example

``` shell
scraperlite https://go.dev \
    whyGo.html 'body > header > div > nav > div > ul > li:nth-child(1)' \
    firstEventWhenWhere.txt '#event_slide0 > div.GoCarousel-eventBody > div > div.GoCarousel-eventDate'
```

In a sqlite3 shell:

``` shell
sqlite> select t, json_extract(content, '$.firstEventWhenWhere.txt') as when_where,
  substr(json_extract(content, '$.whyGo.html'), 1, 20) || '...' as why_go_html
  from observations join contents on (id=content_id)
  order by t;
+----------------------------------+-------------------------------+-------------------------+
|                t                 |          when_where           |       why_go_html       |
+----------------------------------+-------------------------------+-------------------------+
| 2022-02-20 14:19:34.115801-04:00 | Feb 21, 2022 | Graz,  Austria | <li class="Header-me... |
+----------------------------------+-------------------------------+-------------------------+
```
