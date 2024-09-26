import { faker } from "@faker-js/faker";

type Blog = {
  id: number;
  title: string;
  content: string;
};

export const blogs = (() => {
  const arr: Blog[] = [];
  for (let i = 0; i < 10000; i++) {
    arr.push({
      id: i + 1,
      title: faker.lorem.words(10),
      content: faker.lorem.paragraphs(3),
    });
  }
  return arr;
})();
