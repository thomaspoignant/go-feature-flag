import { MdOutlineCheck } from "react-icons/md";
import React from "react";
import Link from "@docusaurus/Link";
import { BsBuildings } from "react-icons/bs";
import { RiOpenSourceFill } from "react-icons/ri";

export default function Pricing() {
  return (
    <div className="relative isolate px-6 py-6 sm:py-12 lg:px-8">
      <div
        className="absolute inset-x-0 -top-3 -z-10 transform-gpu overflow-hidden px-36 blur-3xl"
        aria-hidden="true"
      >
        <div className="mx-auto aspect-[1155/678] w-[72.1875rem] bg-gradient-to-tr from-[#000] to-[#c0f2e7] opacity-30"></div>
      </div>
      <div className="mx-auto max-w-2xl text-center lg:max-w-4xl">
        <p className="mt-2 text-4xl font-bold tracking-tight text-gray-900 sm:text-5xl dark:text-blue-100">
          GO Feature Flag is free 🙌
        </p>
      </div>
      <p className="mx-auto mt-6 max-w-2xl text-center text-lg leading-8 text-gray-600 dark:text-gray-300">
        We are 100% OpenSource and we want to continue like that, but we offer
        other services if you need it.
      </p>
      <div className="mx-auto mt-16 grid max-w-lg grid-cols-1 items-center gap-y-6 sm:mt-20 sm:gap-y-0 lg:max-w-4xl lg:grid-cols-2">
        <div className="rounded-3xl rounded-t-3xl bg-white p-8 ring-1 ring-gray-900/10 sm:mx-8 sm:rounded-b-none sm:p-10 lg:mx-0 lg:rounded-bl-3xl lg:rounded-tr-none">
          <h3
            id="tier-hobby"
            className="text-base font-semibold leading-7 text-indigo-600"
          >
            <RiOpenSourceFill className={"w-10 h-10"} />
            <br />
            Opensource
          </h3>
          <p className="mt-4 flex items-baseline gap-x-2">
            <span className="text-5xl font-bold tracking-tight text-gray-900">
              $0
            </span>
            <span className="text-base text-gray-400">/forever</span>
          </p>
          <p className="mt-6 text-base leading-7 text-gray-600">
            The perfect plan if you want to do everything by your-self.
          </p>
          <ul
            role="list"
            className="mt-8 space-y-3 text-sm leading-6 text-gray-600 sm:mt-10"
          >
            <li className="flex gap-x-3">
              <MdOutlineCheck className={"w-6 h-6 text-goff-500"} />
              Access to all features
            </li>
            <li className="flex gap-x-3">
              <MdOutlineCheck className={"w-6 h-6 text-goff-500"} />
              Use in your own infrastructure
            </li>
            <li className="flex gap-x-3">
              <MdOutlineCheck className={"w-6 h-6 text-goff-500"} />
              Unlimited feature flags
            </li>
            <li className="flex gap-x-3">
              <MdOutlineCheck className={"w-6 h-6 text-goff-500"} />
              Community support only
            </li>
          </ul>
          <Link
            to={"/docs/getting-started"}
            className="mt-8 block rounded-md px-3.5 py-2.5 text-center text-sm font-semibold text-indigo-600 ring-1 ring-inset ring-indigo-200 hover:ring-indigo-300 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 sm:mt-10"
          >
            Get started today
          </Link>
          <Link
            to={"https://github.com/sponsors/thomaspoignant"}
            className="mt-1 block rounded-md px-3.5 py-2.5 text-center text-sm font-semibold text-indigo-600 ring-1 ring-inset ring-indigo-200 hover:ring-indigo-300 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 sm:mt-1"
          >
            Sponsor us ❤️
          </Link>
        </div>
        <div className="relative rounded-3xl bg-gray-900 p-8 shadow-2xl ring-1 ring-gray-900/10 sm:p-10">
          <h3
            id="tier-enterprise"
            className="text-base font-semibold leading-7 text-indigo-400"
          >
            <BsBuildings className={"w-10 h-10"} />
            <br />
            Enterprise
          </h3>
          <p className="mt-4 flex items-baseline gap-x-2">
            <span className="text-5xl font-bold tracking-tight text-white">
              Enterprise support
            </span>
          </p>
          <p className="mt-6 text-base leading-7 text-gray-300">
            The perfect plan if you need a bit more help. We are here for you.
          </p>
          <ul
            role="list"
            className="mt-8 space-y-3 text-sm leading-6 text-gray-300 sm:mt-10"
          >
            <li className="flex gap-x-3">
              <MdOutlineCheck className={"w-6 h-6 text-goff-500"} />
              Same as Opensource
            </li>
            <li className="flex gap-x-3">
              <MdOutlineCheck className={"w-6 h-6 text-goff-500"} />
              Premium support
            </li>
            <li className="flex gap-x-3">
              <MdOutlineCheck className={"w-6 h-6 text-goff-500"} />
              SLA on CVE fix
            </li>
            <li className="flex gap-x-3">
              <MdOutlineCheck className={"w-6 h-6 text-goff-500"} />
              Direct communication with the maintainers
            </li>
            <li className="flex gap-x-3">
              <MdOutlineCheck className={"w-6 h-6 text-goff-500"} />
              Help during your integration
            </li>
            <li className="flex gap-x-3">
              <MdOutlineCheck className={"w-6 h-6 text-goff-500"} />
              Preview of the roadmap
            </li>
            <li className="flex gap-x-3">
              <MdOutlineCheck className={"w-6 h-6 text-goff-500"} />
              Training material for your team
            </li>
            <li className="flex gap-x-3">
              <MdOutlineCheck className={"w-6 h-6 text-goff-500"} />
              Maintainer join your slack/teams organisation
            </li>
          </ul>
          <Link
            to={"mailto:contact@gofeatureflag.org?subject=Enterprise support"}
            className=" mt-8 block rounded-md bg-indigo-500 px-3.5 py-2.5 text-center text-sm font-semibold text-white shadow-sm hover:bg-indigo-400 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-500 sm:mt-10"
          >
            Contact us
          </Link>
          <Link
            to={"https://zcal.co/gofeatureflag/30min"}
            className="mt-2 block rounded-md bg-indigo-500 px-3.5 py-2.5 text-center text-sm font-semibold text-white shadow-sm hover:bg-indigo-400 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-500 sm:mt-2"
          >
            Book a meeting 📅
          </Link>
        </div>
      </div>
    </div>
  );
}
