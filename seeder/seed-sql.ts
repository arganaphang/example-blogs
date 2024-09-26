import postgres from "postgres";
import { blogs } from "./data";

async function main() {
  const sql = postgres(
    "postgres://postgres:postgres@localhost:5432/database?sslmode=disable"
  );

  await sql`DELETE FROM blogs`;
  await sql`INSERT INTO blogs ${sql(blogs, "id", "title", "content")}`;
  await sql.end();
}

main();
