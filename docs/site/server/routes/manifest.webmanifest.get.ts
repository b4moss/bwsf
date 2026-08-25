export default defineEventHandler((event) => {
  setResponseStatus(event, 404);
  setHeader(event, "content-type", "application/manifest+json");
  return "{}";
});
