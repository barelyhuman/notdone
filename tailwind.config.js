/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./templates/**/*.{html,js}"],
  theme: {
    extend: {},
  },
  daisyui: {
    themes: ["cupcake"],
  },
  plugins: [require("daisyui")],
};
