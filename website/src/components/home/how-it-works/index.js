import React from 'react';
import Link from '@docusaurus/Link';

export function HowItWorks() {
  return (
    <section
      className="py-[3rem] text-center"
      aria-labelledby="how-it-works-title">
      <span
        id="how-it-works-title"
        className="text-gray-800 dark:text-gray-50 text-5xl font-poppins font-bold tracking-[-0.18rem]">
        How it works?
      </span>
      <p className="mt-3 mb-8 max-w-3xl mx-auto text-gray-500 dark:text-gray-400 text-[1.05rem] leading-relaxed">
        <p>
          One <strong>relay-proxy</strong> on your infrastructure, your
          applications consume flags through OpenFeature SDKs, and your flag
          configuration lives in a file you already control.
        </p>{' '}
        <p>
          <strong>Retrievers</strong>, <strong>notifiers</strong>, and{' '}
          <strong>exporters</strong> plug into the stack you are using.
        </p>
      </p>
      <div className="max-w-5xl mx-auto px-4">
        <div className="rounded-2xl bg-white dark:bg-white/95 p-6 md:p-8 shadow-sm border border-solid border-[#273437]/10">
          <img
            src="/docs/openfeature/architecture.svg"
            alt="GO Feature Flag architecture: OpenFeature SDKs talking to the relay-proxy, which loads flag configuration via retrievers and emits events via notifiers and exporters."
            className="w-full h-auto"
            loading="lazy"
          />
        </div>
        <p className="mt-6 text-gray-500 dark:text-gray-400 text-[1.05rem] leading-relaxed">
          Want the full picture? Read the{' '}
          <Link to="/docs/concepts/architecture">
            architecture documentation
          </Link>{' '}
          for a deeper dive into every component.
        </p>
      </div>
    </section>
  );
}
