import React from 'react';
import PropTypes from 'prop-types';
import lyft from '@site/static/img/using-it/lyft.png';
import tencent from '@site/static/img/using-it/tencent.png';
import minder from '@site/static/img/using-it/minder.png';
import castai from '@site/static/img/using-it/castai.png';
import grafana from '@site/static/img/using-it/grafana.png';
import alternativepayment from '@site/static/img/using-it/alternativepayment.png';
import agentero from '@site/static/img/using-it/agentero.png';
import mecena from '@site/static/img/using-it/mecena.png';
import landsend from '@site/static/img/using-it/landsend.png';
import helloworldcs from '@site/static/img/using-it/helloworldcs.png';

export function UsingIt() {
  const companies = [
    {
      name: 'Grafana Labs',
      logo: grafana,
      url: 'https://grafana.com',
      imgClassName: 'max-w-48',
    },
    {
      name: 'Cast.ai',
      logo: castai,
      url: 'https://cast.ai',
      imgClassName: 'max-w-24',
    },
    {
      name: 'Lyft',
      logo: lyft,
      url: 'https://lyft.com',
      imgClassName: 'max-w-16',
    },
    {
      name: 'Tencent',
      logo: tencent,
      url: 'https://tencent.com',
    },
    {
      name: 'Minder',
      logo: minder,
      url: 'https://github.com/mindersec/minder',
    },
    {
      name: 'Agentero',
      logo: agentero,
      url: 'https://agentero.com/',
    },
    {
      name: "Lands'end",
      logo: landsend,
      url: 'https://www.landsend.com/',
    },
    {
      name: 'Alternative Payments',
      logo: alternativepayment,
      url: 'https://www.alternativepayments.io/',
    },
    {
      name: 'mecena',
      logo: mecena,
      url: 'https://mecena.co/',
    },
    {
      name: 'HelloWorld CS',
      logo: helloworldcs,
      url: 'https://helloworldcs.org/',
    },
  ];

  return (
    <section className={'pt-5 px-5'}>
      <div className="grid grid-pad text-center">
        <span className="text-3xl">Trusted by developers from</span>
        <UsingItLogos companies={companies} />
      </div>
    </section>
  );
}

UsingItLogos.propTypes = {
  companies: PropTypes.arrayOf(
    PropTypes.shape({
      url: PropTypes.string.isRequired,
      logo: PropTypes.string.isRequired,
      name: PropTypes.string.isRequired,
      imgClassName: PropTypes.string,
    })
  ).isRequired,
};

function UsingItLogos({companies}) {
  return (
    <div
      className={
        'grid grid-cols-1 sm:grid-cols-2 md:grid-cols-5 2xl:grid-cols-10 mt-8 items-center mb-0'
      }>
      {companies.map(company => (
        <div key={company.name}>
          <img
            src={company.logo}
            alt={company.name}
            className={company.imgClassName ? company.imgClassName : 'max-w-36'}
          />
        </div>
      ))}
    </div>
  );
}
