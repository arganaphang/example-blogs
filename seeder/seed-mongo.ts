import * as mongoose from "mongoose";
import { blogs } from "./data";

const blogSchema = new mongoose.Schema(
  {
    title: { type: String, required: true },
    content: { type: String, required: true },
    created_at: { type: Date, required: true, default: Date.now },
    updated_at: { type: Date, required: true, default: Date.now },
  },
  {}
);

export type Blog = mongoose.InferSchemaType<typeof blogSchema>;
export const Blog = mongoose.model("Blog", blogSchema);

async function main() {
  try {
    await mongoose.connect("mongodb://mongo:mongo@localhost:27017/");
    await Blog.collection.deleteMany();
    await Blog.collection.insertMany(
      blogs.map(
        (blog) => new Blog({ title: blog.title, content: blog.content })
      )
    );
  } catch (error) {
    console.log(error);
  } finally {
    await mongoose.disconnect();
  }
}

main();
