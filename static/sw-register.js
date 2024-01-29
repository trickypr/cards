// @ts-check
if ("serviceWorker" in navigator) {
  try {
    const registration = await navigator.serviceWorker.register(
      "/static/sw.js",
      { scope: "/" },
    );
  } catch {}
}
