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
