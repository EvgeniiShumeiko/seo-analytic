// Ищет технологии по JS-объектам страницы после рендеринга.
// Возвращает массив {name, category, version}; выполняется в контексте страницы через chromedp.
(() => {
  const found = [];
  const add = (name, category, version) =>
    found.push({ name, category, version: version ? String(version) : "" });
  const w = window;

  // Фреймворки вешают служебные свойства на DOM-узлы: React — __reactFiber$…/__reactContainer$…,
  // Vue 3 — __vue_app__, Vue 2 — __vue__. Достаточно проверить корень и первый уровень body.
  const nodes = [
    document.documentElement,
    document.body,
    ...document.querySelectorAll("body > *, #root, #app, #__next, #__nuxt"),
  ].filter(Boolean);
  const hasNodeKey = (prefix) => nodes.some((n) => Object.keys(n).some((k) => k.startsWith(prefix)));
  const hasGlobal = (prefix) => Object.keys(w).some((k) => k.startsWith(prefix));

  if (hasNodeKey("__react") || w.React) add("React", "Фронтенд", w.React?.version);
  const vueApp = nodes.find((n) => n.__vue_app__)?.__vue_app__;
  if (vueApp || hasNodeKey("__vue__") || w.Vue) add("Vue.js", "Фронтенд", vueApp?.version || w.Vue?.version);
  const ngEl = document.querySelector("[ng-version]");
  if (ngEl || w.ng) add("Angular", "Фронтенд", ngEl?.getAttribute("ng-version"));
  if (w.angular?.version) add("AngularJS", "Фронтенд", w.angular.version.full);
  if (hasGlobal("__sveltekit")) add("SvelteKit", "Фронтенд");
  if (w.next || w.__NEXT_DATA__) add("Next.js", "Фронтенд", w.next?.version);
  if (w.__NUXT__ || w.$nuxt) add("Nuxt", "Фронтенд");
  if (w.___gatsby) add("Gatsby", "Фронтенд");
  if (w.__remixContext) add("Remix", "Фронтенд");
  if (w.Ember) add("Ember.js", "Фронтенд", w.Ember.VERSION);
  if (w.Backbone) add("Backbone.js", "Фронтенд", w.Backbone.VERSION);
  if (w.Alpine) add("Alpine.js", "Фронтенд", w.Alpine.version);
  if (w.htmx) add("htmx", "Фронтенд", w.htmx.version);

  if (w.jQuery?.fn?.jquery) add("jQuery", "JS-библиотека", w.jQuery.fn.jquery);
  if (w._?.VERSION) add("Lodash/Underscore", "JS-библиотека", w._.VERSION);
  if (w.moment?.version) add("Moment.js", "JS-библиотека", w.moment.version);
  if (w.__APOLLO_CLIENT__ || w.__APOLLO_STATE__) add("Apollo GraphQL", "JS-библиотека");
  if (hasGlobal("webpackChunk")) add("webpack", "Сборщик");
  if (w.Sentry) add("Sentry", "Мониторинг ошибок", w.Sentry.SDK_VERSION);

  if (w.BX) add("1С-Битрикс", "CMS");
  if (w.wp) add("WordPress", "CMS");
  if (w.Shopify) add("Shopify", "E-commerce");
  if (w.t_onReady) add("Tilda", "Конструктор");

  return found;
})()
