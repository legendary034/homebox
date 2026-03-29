export default defineNuxtRouteMiddleware(() => {
  const api = useUserApi();
  const ctx = useAuthContext();

  // Fetch real user data in the background without blocking the boot process.
  api.user.self().then(({ data }) => {
    if (data && data.item) {
      ctx.user = data.item;
    }
  }).catch(e => {
    console.warn("Background user fetch failed (expected if DB is empty):", e);
  });
});
