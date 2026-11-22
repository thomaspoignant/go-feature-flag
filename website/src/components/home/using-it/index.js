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
      mobileClassName: 'h-5',
      logo: grafana,
      url: 'https://grafana.com',
      imgClassName: 'lg:max-w-48',
    },
    {
      name: 'Miro',
      mobileClassName: 'h-8',
      logo: miro,
      url: 'https://miro.com',
      imgClassName: 'lg:max-w-32',
    },
    {
      name: 'Lyft',
      mobileClassName: 'h-10',
      logo: lyft,
      url: 'https://lyft.com',
      imgClassName: 'lg:max-w-12',
    },
    {
      name: 'Tencent',
      mobileClassName: 'h-4',
      logo: tencent,
      url: 'https://tencent.com',
    },
    {
      name: 'Cast.ai',
      mobileClassName: 'h-8',
      logo: castai,
      url: 'https://cast.ai',
      imgClassName: 'lg:max-w-24',
    },
    {
      name: "Lands'end",
      mobileClassName: 'h-4',
      logo: landsend,
      url: 'https://www.landsend.com/',
    },
    {
      name: 'Solaris',
      mobileClassName: 'h-6',
      logo: solaris,
      url: 'https://www.solarisgroup.com/',
      imgClassName: 'lg:max-w-32',
    },
    {
      name: 'Kargo',
      mobileClassName: 'h-7',
      logo: kargo,
      url: 'https://www.kargo.com/',
      imgClassName: 'lg:max-w-20',
    },
    {
      name: 'DataGalaxy',
      mobileClassName: 'h-5',
      logo: datagalaxy,
      url: 'https://www.datagalaxy.com/',
      imgClassName: 'lg:max-w-36',
    },
    {
      name: 'Minder',
      mobileClassName: 'h-6',
      logo: minder,
      url: 'https://github.com/mindersec/minder',
    },
    {
      name: 'Alternative Payments',
      mobileClassName: 'h-6.5',
      logo: alternativepayment,
      url: 'https://www.alternativepayments.io/',
    },
    {
      name: 'Agentero',
      mobileClassName: 'h-7',
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
      mobileClassName: PropTypes.string,
    })
  ).isRequired,
};

function UsingItLogos({companies}) {
  return (
    <div
      className={
        'grid grid-cols-2 gap-x-6 gap-y-6 max-w-sm mx-auto mt-8 mb-8 items-center ' +
        'lg:grid-cols-6 lg:gap-0 lg:max-w-none lg:mx-0 lg:mb-0'
      }>
      {companies.map(company => (
        <div
          key={company.name}
          className={
            'flex items-center justify-center h-10 lg:block lg:h-auto'
          }>
          <img
            src={company.logo}
            alt={company.name}
            className={[
              // Mobile: height normalized per logo so they share the same visual weight.
              company.mobileClassName ?? 'h-6',
              'w-auto max-w-full object-contain',
              // Desktop: keep the original sizing.
              'lg:h-auto',
              company.imgClassName ? company.imgClassName : 'lg:max-w-28',
            ].join(' ')}
          />
        </div>
      ))}
    </div>
  );
}
