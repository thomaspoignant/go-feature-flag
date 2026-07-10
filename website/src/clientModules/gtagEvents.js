/**
 * Global GA4 conversion event tracking.
 *
 * Registered as a Docusaurus client module (see `clientModules` in
 * docusaurus.config.js). The gtag plugin exposes `window.gtag` in production
 * builds; in `npm start` dev it is undefined, so `sendEvent` is a safe no-op.
 *
 * A single delegated `document` click listener emits four key events:
 *  - `github_click`         any link to the GitHub repo (site-wide, URL pattern)
 *  - `slack_join`           any Slack link (site-wide, URL pattern)
 *  - `install`              the home hero "Get Started" CTA (data-ga-event marker)
 *  - `relay_proxy_download` the home QuickStart Docker copy button (marker)
 *
 * Site-wide events use URL matching because those links live in many places,
 * including raw-HTML <a> strings inside docusaurus.config.js that a React
 * onClick handler cannot reach. Home-page-specific events use explicit
 * `data-ga-event` attributes on the exact elements.
 */
import ExecutionEnvironment from '@docusaurus/ExecutionEnvironment';

function sendEvent(name, params) {
  if (!name) return;
  if (typeof window.gtag === 'function') {
    window.gtag('event', name, params);
  }
}

function text(el) {
  return (el.textContent || '').trim().slice(0, 100);
}

if (ExecutionEnvironment.canUseDOM) {
  document.addEventListener(
    'click',
    e => {
      // Click targets are Elements in modern browsers, but guard defensively:
      // a text-node target has no `closest`, so fall back to its parent element.
      const target =
        e.target instanceof Element ? e.target : e.target?.parentElement;
      if (!target) return;

      // (a) Explicit markers: install CTA + relay-proxy copy button.
      const marked = target.closest('[data-ga-event]');
      if (marked) {
        sendEvent(marked.dataset.gaEvent, {
          method: marked.dataset.gaMethod,
          link_text: text(marked),
        });
        return;
      }

      // (b) URL-pattern links: github_click + slack_join.
      const a = target.closest('a[href]');
      if (!a) return;
      let url;
      try {
        url = new URL(a.href, window.location.origin);
      } catch {
        return;
      }

      // Slack: internal /slack redirect OR Gophers invite/workspace.
      if (
        url.pathname === '/slack' ||
        url.hostname === 'gophers.slack.com' ||
        url.hostname === 'invite.slack.golangbridge.org'
      ) {
        sendEvent('slack_join', {
          method: url.pathname === '/slack' ? 'internal_link' : 'slack_invite',
          link_url: url.href,
          link_text: text(a),
        });
        return;
      }

      // GitHub repo (exclude docs "edit this page" source links to cut noise).
      if (
        url.hostname === 'github.com' &&
        url.pathname.startsWith('/thomaspoignant/go-feature-flag') &&
        !/\/(tree|blob|edit)\//.test(url.pathname)
      ) {
        sendEvent('github_click', {
          link_url: url.href,
          link_text: text(a),
        });
      }
    },
    true // capture phase — fires before client-side navigation
  );
}
