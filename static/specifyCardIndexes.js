const update = (el, name, detail) => {
  el.dataset[name] = detail;
  el.dispatchEvent(new CustomEvent(name, { detail }));
};

window.specifyCardIndexes = () => {
  let last = 0;
  document
    .querySelectorAll("#learn-cards > .card__container")
    .forEach((element, index) => {
      last = index;
      update(element, "index", index);
    });

  const learn = document.getElementById("learn-cards");
  update(learn, "last", last);
};
window.specifyCardIndexes();

// @ts-check

const stats = [1, 2, 3, 4, 5];
const total = stats.reduce((a, b) => a + b, 0);
const colors = ["red", "yellow", "green", "blue", "orange"];

/**
 * @typedef {object} out
 * @property {number} before
 * @property {string[]} vals
 */

/**
 * @param {out} input
 * @param {number} count
 * @param {number} index
 * @return {out}
 */
const reduceFn = ({ before, vals }, count, index) => ({
  before: before + count,
  vals: [...vals, `${colors[index]} ${before}% ${before + count}%`],
});

const grad = stats.reduce(reduceFn, { before: 0, vals: [] }).vals.join(", ");
