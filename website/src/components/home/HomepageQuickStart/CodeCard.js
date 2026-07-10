import React, {useState} from 'react';
import clsx from 'clsx';
import PropTypes from 'prop-types';
import {Highlight} from 'prism-react-renderer';
import {usePrismTheme} from '@docusaurus/theme-common';
import Link from '@docusaurus/Link';
import styles from './CodeCard.module.css';

function FileIcon({className}) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.75"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      className={className}>
      <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
      <polyline points="14 2 14 8 20 8" />
    </svg>
  );
}

FileIcon.propTypes = {
  className: PropTypes.string,
};

function CopyIcon({className}) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.75"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      className={className}>
      <rect width="13" height="13" x="9" y="9" rx="2" />
      <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
    </svg>
  );
}

CopyIcon.propTypes = {
  className: PropTypes.string,
};

function CheckIcon({className}) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      className={className}>
      <polyline points="20 6 9 17 4 12" />
    </svg>
  );
}

CheckIcon.propTypes = {
  className: PropTypes.string,
};

function CopyButton({code, analyticsEvent, analyticsMethod}) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(code);
      setCopied(true);
      setTimeout(() => setCopied(false), 1800);
    } catch (e) {
      // clipboard write can fail in insecure contexts; ignore silently
    }
  };

  return (
    <button
      type="button"
      onClick={handleCopy}
      aria-label="Copy code to clipboard"
      {...(analyticsEvent && {'data-ga-event': analyticsEvent})}
      {...(analyticsMethod && {'data-ga-method': analyticsMethod})}
      className={clsx(
        'flex items-center gap-1.5 px-3 py-1.5 rounded-md text-xs font-poppins font-medium',
        'text-gray-700 dark:text-gray-200',
        'bg-gray-100 hover:bg-gray-200 dark:bg-[#2a2a2a] dark:hover:bg-[#363636]',
        'transition-colors focus:outline-none focus:ring-2 focus:ring-[var(--ifm-color-primary)]'
      )}>
      {copied ? (
        <CheckIcon className="w-3.5 h-3.5 text-[var(--ifm-color-primary)]" />
      ) : (
        <CopyIcon className="w-3.5 h-3.5" />
      )}
      <span>{copied ? 'Copied' : 'Copy'}</span>
    </button>
  );
}

CopyButton.propTypes = {
  code: PropTypes.string.isRequired,
  analyticsEvent: PropTypes.string,
  analyticsMethod: PropTypes.string,
};

function CodeBody({code, language}) {
  const prismTheme = usePrismTheme();
  return (
    <Highlight code={code} language={language} theme={prismTheme}>
      {({className, style, tokens, getLineProps, getTokenProps}) => {
        let lineStartOffset = 0;
        return (
          <pre
            className={clsx(
              className,
              styles.codePre,
              'm-0 px-5 py-4 text-[0.85rem] leading-6 font-mono'
            )}
            style={{
              ...style,
              background: 'transparent',
              overflowX: 'auto',
              overflowY: 'hidden',
            }}>
            {tokens.map(line => {
              const lineOffset = lineStartOffset;
              const lineText = line.map(t => t.content).join('');
              lineStartOffset += lineText.length;
              let tokenOrdinal = 0;
              return (
                <div key={`line-${lineOffset}`} {...getLineProps({line})}>
                  {line.map(token => {
                    const typesKey = (token.types ?? []).join('.');
                    const tokenKey = `tok-${lineOffset}-${tokenOrdinal}-${typesKey}`;
                    tokenOrdinal += 1;
                    return <span key={tokenKey} {...getTokenProps({token})} />;
                  })}
                </div>
              );
            })}
          </pre>
        );
      }}
    </Highlight>
  );
}

CodeBody.propTypes = {
  code: PropTypes.string.isRequired,
  language: PropTypes.string.isRequired,
};

export function CodeCard({
  filename,
  language,
  code,
  tabs,
  callout,
  moreLink,
  analyticsEvent,
  analyticsMethod,
}) {
  const isTabbed = Array.isArray(tabs) && tabs.length > 0;
  const [activeValue, setActiveValue] = useState(
    isTabbed ? tabs[0].value : null
  );

  const activeTab = isTabbed
    ? (tabs.find(t => t.value === activeValue) ?? tabs[0])
    : null;
  const currentCode = (isTabbed ? activeTab.code : code) ?? '';
  const currentLang = isTabbed ? activeTab.language : language;
  const currentName = isTabbed
    ? (activeTab.filename ?? activeTab.label)
    : filename;

  return (
    <div
      className={clsx(
        'rounded-xl overflow-hidden text-left shadow-lg',
        'border-2 border-gray-400 dark:border-[#71717a]',
        'bg-white dark:bg-[#1a1a1a]'
      )}>
      {isTabbed && (
        <div
          role="tablist"
          aria-label="Language"
          className={clsx(
            styles.tabList,
            'flex items-center gap-1 px-3 pt-2 overflow-x-auto border-b border-gray-200 dark:border-[#262626]'
          )}>
          {tabs.map(t => {
            const isActive = t.value === activeTab.value;
            return (
              <button
                key={t.value}
                type="button"
                role="tab"
                aria-selected={isActive}
                onClick={() => setActiveValue(t.value)}
                className={clsx(
                  'px-3 pt-1.5 pb-2 -mb-px text-sm font-medium font-poppins whitespace-nowrap',
                  'bg-transparent border-0 border-b-2 border-solid transition-colors',
                  'focus:outline-none focus-visible:outline-none',
                  isActive
                    ? 'border-b-[var(--ifm-color-primary)] text-[var(--ifm-color-primary)]'
                    : 'border-b-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'
                )}>
                {t.label}
              </button>
            );
          })}
          {moreLink && (
            <Link
              to={moreLink.to}
              aria-label={moreLink.ariaLabel ?? 'See all SDKs'}
              className={clsx(
                'flex items-center gap-1 px-3 pt-1.5 pb-2 -mb-px',
                'text-sm font-medium font-poppins whitespace-nowrap no-underline hover:no-underline',
                'border-0 border-b-2 border-solid border-b-transparent',
                'text-gray-500 hover:text-[var(--ifm-color-primary)] dark:text-gray-400 dark:hover:text-[var(--ifm-color-primary)]',
                'transition-colors'
              )}>
              <span>{moreLink.label ?? 'More'}</span>
              <svg
                xmlns="http://www.w3.org/2000/svg"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
                aria-hidden="true"
                className="w-3.5 h-3.5">
                <polyline points="9 18 15 12 9 6" />
              </svg>
            </Link>
          )}
        </div>
      )}
      <div
        className={clsx(
          'flex items-center justify-between gap-3 px-3 py-2',
          !isTabbed && 'border-b border-gray-100 dark:border-[#262626]'
        )}>
        <div className="flex items-center gap-2 pl-1 text-gray-500 dark:text-gray-400">
          {isTabbed ? (
            <span className="flex items-center w-4 h-4 [&>svg]:w-full [&>svg]:h-full">
              {activeTab.icon ?? <FileIcon className="w-4 h-4 opacity-80" />}
            </span>
          ) : (
            <FileIcon className="w-4 h-4 opacity-80" />
          )}
          <span className="font-mono text-sm">
            {isTabbed
              ? (activeTab.displayName ?? activeTab.language)
              : currentName}
          </span>
        </div>
        <CopyButton
          code={currentCode}
          analyticsEvent={analyticsEvent}
          analyticsMethod={analyticsMethod}
        />
      </div>
      <CodeBody code={currentCode} language={currentLang} />
      {callout && (
        <div className="flex items-start gap-3 px-4 py-3 border-t border-gray-100 dark:border-[#262626] bg-gray-50/80 dark:bg-[#161616] text-sm text-gray-700 dark:text-gray-200">
          <span
            aria-hidden="true"
            className="mt-1.5 inline-block w-2 h-2 rounded-full bg-[var(--ifm-color-primary)] shrink-0"
          />
          <p className="m-0 leading-relaxed text-left">{callout}</p>
        </div>
      )}
    </div>
  );
}

CodeCard.propTypes = {
  filename: PropTypes.string,
  language: PropTypes.string,
  code: PropTypes.string,
  tabs: PropTypes.arrayOf(
    PropTypes.shape({
      value: PropTypes.string.isRequired,
      label: PropTypes.string.isRequired,
      language: PropTypes.string.isRequired,
      code: PropTypes.string.isRequired,
      filename: PropTypes.string,
      displayName: PropTypes.string,
      icon: PropTypes.node,
    })
  ),
  callout: PropTypes.node,
  moreLink: PropTypes.shape({
    to: PropTypes.string.isRequired,
    label: PropTypes.string,
    ariaLabel: PropTypes.string,
  }),
  analyticsEvent: PropTypes.string,
  analyticsMethod: PropTypes.string,
};
