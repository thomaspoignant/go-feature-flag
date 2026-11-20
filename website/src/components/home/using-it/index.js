import React from 'react';
import PropTypes from 'prop-types';
import lyft from '@site/static/img/using-it/lyft.png';
import tencent from '@site/static/img/using-it/tencent.png';
import minder from '@site/static/img/using-it/minder.png';
import castai from '@site/static/img/using-it/castai.png';
import grafana from '@site/static/img/using-it/grafana.png';
import alternativepayment from '@site/static/img/using-it/alternativepayment.png';
import agentero from '@site/static/img/using-it/agentero.png';
import landsend from '@site/static/img/using-it/landsend.png';
import miro from '@site/static/img/using-it/miro.png';
import solaris from '@site/static/img/using-it/solaris.png';
import kargo from '@site/static/img/using-it/kargo.png';
import datagalaxy from '@site/static/img/using-it/datagalaxy.png';

export function UsingIt() {
  const companies = [
    {
      name: 'Grafana Labs',
      logo: grafana,
      url: 'https://grafana.com',
      imgClassName: 'max-w-48',
    },
    {
      name: 'Miro',
      logo: miro,
      url: 'https://miro.com',
      imgClassName: 'max-w-32',
    },
    {
      name: 'Lyft',
      logo: lyft,
      url: 'https://lyft.com',
      imgClassName: 'max-w-12',
    },
    {
      name: 'Tencent',
      logo: tencent,
      url: 'https://tencent.com',
    },
    {
      name: 'Cast.ai',
      logo: castai,
      url: 'https://cast.ai',
      imgClassName: 'max-w-24',
    },
    {
      name: "Lands'end",
      logo: landsend,
      url: 'https://www.landsend.com/',
    },
    {
      name: 'Solaris',
      logo: solaris,
      url: 'https://www.solarisgroup.com/',
      imgClassName: 'max-w-32',
    },
    {
      name: 'Kargo',
      logo: kargo,
      url: 'https://www.kargo.com/',
      imgClassName: 'max-w-20',
    },
    {
      name: 'DataGalaxy',
      logo: datagalaxy,
      url: 'https://www.datagalaxy.com/',
      imgClassName: 'max-w-36',
    },
    {
      name: 'Minder',
      logo: minder,
      url: 'https://github.com/mindersec/minder',
    },
    {
      name: 'Alternative Payments',
      logo: alternativepayment,
      url: 'https://www.alternativepayments.io/',
    },
    {
      name: 'Agentero',
      logo: agentero,
      url: 'https://agentero.com/',
    },
  ];

  return (
    <section className={'pt-5 px-5 bg-white dark:bg-[#1b1b1d]'}>
      <div className="grid grid-pad text-center">
        <span className="text-2xl">
          Trusted in production by engineering teams at
        </span>
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
    <div className={'grid grid-cols-2 lg:grid-cols-6 mt-8 items-center mb-0'}>
      {companies.map(company => (
        <div key={company.name}>
          <img
            src={company.logo}
            alt={company.name}
            className={company.imgClassName ? company.imgClassName : 'max-w-28'}
          />
        </div>
      ))}
    </div>
  );
}
