import { forwardRef } from 'react';

export const TavilySVG = forwardRef<SVGSVGElement, React.SVGProps<SVGSVGElement>>((props, ref) => (
  <svg
    ref={ref}
    width="24"
    height="24"
    viewBox="0 0 24 24"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    {...props}
  >
    <g id="tavily">
      <g id="Icon">
        <path
          transform="scale(0.55)"
          d="m16.44.964 4.921 7.79c.79 1.252-.108 2.883-1.588 2.883H17.76V23.3h-2.91V.088c.61 0 1.22.292 1.59.876z"
          fill="#8FBCFA" />
        <path
          transform="scale(0.55)"
          d="M8.342 8.755 13.263.964a1.864 1.864 0 0 1 1.59-.876V23.3a4.87 4.87 0 0 0-.252-.006c-.99 0-1.907.311-2.658.842V11.637H9.93c-1.48 0-2.38-1.631-1.589-2.882z"
          fill="#468BFF" />
        <path
          transform="scale(0.55)"
          d="M30.278 31H18.031a4.596 4.596 0 0 0 1.219-2.91h22.577c0 .61-.292 1.22-.875 1.59L33.16 34.6c-1.251.791-2.883-.108-2.883-1.588V31z"
          fill="#FDBB11" />
        <path
          transform="scale(0.55)"
          d="m33.16 21.581 7.79 4.921c.585.369.876.979.876 1.589H19.25a4.619 4.619 0 0 0-.858-2.91h11.887V23.17c0-1.48 1.631-2.38 2.882-1.589z"
          fill="#F6D785" />
        <path
          transform="scale(0.55)"
          d="m8.24 34.25-7.107 7.108a1.864 1.864 0 0 0 1.742.504l8.989-2.03c1.443-.325 1.961-2.114.915-3.16l-1.423-1.423 5.356-5.356a2.805 2.805 0 0 0 0-3.966l-.074-.075L8.24 34.25z"
          fill="#FF9A9D" />
        <path
          transform="scale(0.55)"
          d="m7.243 31.135 5.355-5.356a2.805 2.805 0 0 1 3.967 0l.074.074-8.397 8.397-7.108 7.108a1.864 1.864 0 0 1-.504-1.742l2.029-8.989c.325-1.444 2.115-1.961 3.161-.915l1.423 1.423z"
          fill="#FE363B" />
      </g>
    </g>
  </svg>
));

TavilySVG.displayName = 'TavilySVG';