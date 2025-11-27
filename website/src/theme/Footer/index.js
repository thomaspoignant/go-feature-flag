import React from 'react';
import {useThemeConfig} from '@docusaurus/theme-common';
import FooterLayout from './Layout';
import FooterLinks from './Links';
import FooterLogo from './Logo';
import FooterCopyright from './Copyright';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import useIsBrowser from '@docusaurus/useIsBrowser';
import MailchimpSubscribe from 'react-mailchimp-subscribe';
import styles from './styles.module.css';
import clsx from 'clsx';

function Footer() {
  const {footer} = useThemeConfig();
  if (!footer) {
    return null;
  }
  const {copyright, links, logo, style} = footer;
  const page = useIsBrowser() ? window.location : 'not_reachable';
  const {siteConfig} = useDocusaurusContext();
  const url = `${siteConfig.customFields.mailchimpURL};SIGNUP=${page}`;
  return (
    <FooterLayout
      style={style}
      links={links && links.length > 0 && <FooterLinks links={links} />}
      logo={logo && <FooterLogo logo={logo} />}
      copyright={copyright && <FooterCopyright copyright={copyright} />}
      newsletter={<NewsletterForm url={url} />}
    />
  );
}

const CustomForm = ({status, message, onValidated}) => {
  let email;
  const submit = () =>
    email &&
    email.value.indexOf('@') > -1 &&
    onValidated({
      EMAIL: email.value,
    });

  return (
    <div className={'text-center items-center'}>
      <input
        className={clsx('w-full p-5 rounded rounded-2xl border-2 max-w-xl')}
        ref={node => (email = node)}
        type="email"
        placeholder="Your Email"
      />
      {status === 'error' && (
        <div
          className="p-4 max-w-xl my-4 text-sm text-red-800 rounded-lg bg-red-50 dark:bg-gray-800 dark:text-red-400"
          role="alert">
          <i className="fa-solid fa-triangle-exclamation pr-4"></i>
          <span className="font-medium">Danger alert!</span> {message}
        </div>
      )}
      {status === 'success' && (
        <div
          className="p-4 max-w-xl mb-4 text-sm text-green-800 rounded-lg bg-green-50 dark:bg-gray-800 dark:text-green-400"
          role="alert">
          <i className="fa-solid fa-check pr-4"></i>
          <span className="font-medium">Success1!</span> {message}
        </div>
      )}
      <button
        type="button"
        className="w-full max-w-xl mt-4 h-12 rounded-2xl cursor-pointer text-white bg-gradient-to-br from-purple-600 to-blue-500 hover:bg-gradient-to-bl focus:ring-4 focus:outline-none focus:ring-blue-300 dark:focus:ring-blue-800 font-medium text-sm px-5 py-2.5 text-center me-2 mb-2"
        onClick={submit}>
        <i className="fa-regular fa-paper-plane"></i> Subscribe
      </button>
    </div>
  );
};

const NewsletterForm = ({url}) => (
  <div>
    <div className="text-center pt-0 mx-4 text-5xl mb-4 text-gray-800 dark:text-gray-100 font-poppins font-[800] tracking-[-0.08rem]">
      <i className="fa-solid fa-envelope"></i> <br />
      Get the latest <br /> Updates
    </div>
    <div
      className={'text-md text-center text-gray-600 dark:text-gray-400 mb-2'}>
      Get all the tips, updates and contents from GO Feature Flag. Your inbox
      will love it <i className="fa-solid fa-envelope"></i>
    </div>
    <MailchimpSubscribe
      url={url}
      render={({subscribe, status, message}) => (
        <CustomForm
          status={status}
          message={message}
          onValidated={formData => subscribe(formData)}
        />
      )}
    />
    <div>
      <p className="text-sm text-center text-gray-600 dark:text-gray-400">
        We will never share your email address. You can unsubscribe at any time.
      </p>
    </div>
  </div>
);

export default React.memo(Footer);
