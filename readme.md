# Blog Daily Randomize using Postgres & MongoDB

## Todos

- [x] Gracefully Shutdown (Close database connection)
- [ ] Update Golang using Goqu
- [ ] Simulate Real case using Cron

## Materialize View

- Create

```sql
CREATE MATERIALIZED VIEW IF NOT EXISTS random_blogs AS select * from blogs ORDER BY RANDOM();
CREATE UNIQUE INDEX ON random_blogs(id);
```

- Refresh
  > Will lock table `blogs` while refreshing

```sql
REFRESH MATERIALIZED VIEW random_blogs;
```

- Refresh Concurrently (not working in this case, donno why)
  > To avoid lock table `blogs` while refreshing
  > Notice that CONCURRENTLY option is only available in PostgreSQL 9.4 or later.

```sql
REFRESH MATERIALIZED VIEW CONCURRENTLY random_blogs;
```

- Drop

```sql
DROP MATERIALIZED VIEW random_blogs;
```

- Another Query to simulate and check count of original table + clone table

```sql
INSERT INTO public.blogs (title, content) VALUES ('another testing', 'test content test');

select count(1) from public.blogs;

select count(1) from public.random_blogs;

REFRESH MATERIALIZED VIEW random_blogs;

with ori as (select count(1) from public.blogs), clone as (select count(1) from public.random_blogs)
select ori, clone from ori, clone;
```
