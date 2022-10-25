import React from 'react';
import Footer from '@theme-original/Footer';
import styles from './styles.module.css'
import MailchimpSubscribe from "react-mailchimp-subscribe"
import clsx from "clsx";
import useDocusaurusContext from "@docusaurus/useDocusaurusContext";
import useIsBrowser from '@docusaurus/useIsBrowser';

export default function FooterWrapper(props) {
  const isBrowser = useIsBrowser();
  const page = isBrowser? window.location : "not_reachable";
  const {siteConfig} = useDocusaurusContext();
  const url = `${siteConfig.customFields.mailchimpURL};SIGNUP=${page}`;

  const CustomForm = ({ status, message, onValidated }) => {
    let email;
    const submit = () =>
      email &&
      email.value.indexOf("@") > -1 &&
      onValidated({
        EMAIL: email.value,
      });

    return (
      <div>
        <input className={styles.emailInput} ref={node => (email = node)} type="email" placeholder="Email" />
        <button className="pushy__btn pushy__btn--df pushy__btn--black" onClick={submit}>
          Subscribe <i className="fa-regular fa-paper-plane"></i>
        </button>
        {status === "error" && (<div className={clsx(styles.newsletterMessage, styles.error)} dangerouslySetInnerHTML={{ __html: message }} />)}
        {status === "success" && (<div className={clsx(styles.newsletterMessage, styles.success)} dangerouslySetInnerHTML={{ __html: message }} />)}
      </div>
    );
  };

  const NewsletterForm = () => (
    <div className={styles.newsletter}>
      <h1>Get the latest GO Feature Flag updates</h1>
      <MailchimpSubscribe
        url={url}
        render={({ subscribe, status, message }) => (
          <CustomForm
            status={status}
            message={message}
            onValidated={formData => subscribe(formData)}
          />
        )}
      />
    </div>
  )

  return (
    <>
      <NewsletterForm />
      <Footer {...props} />
    </>
  );
}
