import React from 'react';

export function Testimonial() {
  return (
    <section>
      <div className="my-20">
        <div className="w-full flex flex-col items-center justify-center gap-2">
          <h1 className="text-4xl text-gray-600 leading-relaxed text-center w-3/5 dark:text-gray-300 italic">
            If your organization is evaluating flag providers, GO Feature Flag
            deserves a serious look, not as &quot;the free option,&quot; but as{' '}
            <span className="text-goff-300">
              a well-engineered piece of infrastructure in its own right
            </span>
          </h1>
          <div className="flex items-center gap-4">
            <div className="rounded-full w-12 h-12 bg-black overflow-hidden">
              <img src="https://github.com/tomflenner.png" />
            </div>
            <div className="flex flex-col tracking-wider">
              <label className="text-gray-600 font-bold text-base dark:text-gray-300">
                Tom Flenner
              </label>
              <label className="text-gray-400 font-normal text-sm">
                Staff Engineer at DataGalaxy
              </label>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
