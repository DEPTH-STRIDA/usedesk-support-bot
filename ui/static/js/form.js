// Глобальные переменные
let organizedCustomSelects;
let tg;
let initData;

// Основная функция, выполняемая при загрузке DOM
document.addEventListener("DOMContentLoaded", function () {
  // Инициализация Telegram
  initializeTg(5, 2000) // 5 попыток с интервалом в 2 секунды
    .then((result) => {
      console.log("Telegram Web App успешно инициализирован");
      tg = result.tg;
      initData = result.initData;

      // Проверяем, что initData не undefined
      if (initData === undefined) {
        console.warn("initData is undefined after initialization");
      } else {
        console.log("initData:", initData);
      }

      // Дальнейшая логика с использованием tg и initData
      organizedCustomSelects = organizeCustomSelects();
      console.log(organizedCustomSelects);

      document
        .getElementById("teg")
        .addEventListener("tegChanged", function (event) {
          console.log(event.detail);
          const tegElement = event.detail.value;
          const tegNumber = extractTegNumber(tegElement);
          if (tegNumber !== null) {
            toggleCustomSelectDisplay(tegNumber);
          } else {
            console.error("Не удалось извлечь номер тега из:", tegElement);
          }
        });

      toggleCustomSelectDisplay(0);

      document.getElementById("send").addEventListener("click", function () {
        setMainButton();
      });

      setupSwipeHandler((swipeInfo) => {
        console.log(`Обнаружен резкий свайп ${swipeInfo.direction}!`);
        console.log(
          `Расстояние: ${
            swipeInfo.distance
          }px, Скорость: ${swipeInfo.speed.toFixed(2)}px/ms, Время: ${
            swipeInfo.time
          }ms`
        );
        if (initData) {
          window.location.href =
            "/admin-menu?initData=" + encodeURIComponent(initData);
        } else {
          console.error("initData is undefined, cannot redirect");
          showAlert(
            "Ошибка: не удалось получить данные инициализации",
            5,
            true
          );
        }
      });
    })
    .catch((error) => {
      console.error("Ошибка инициализации Telegram Web App:", error);
      showAlert("Пожалуйста, запустите форму через телеграмм", 15, true);
    });
});

function handleSubmit() {
  const isEmergency = document.querySelector("#is-emergency").checked;
  const name = document.querySelector("#name").value.trim();
  const place = document.querySelector("#place").value.trim();
  const groupNumber = document.querySelector("#group-number").value.trim();
  const departament = document.querySelector("#teg").textContent;
  const readyProblem = document.querySelector(
    ".problem .custom-select[style*='display: flex'] .select-selected"
  ).textContent;
  const customProblem = document.querySelector("#custom-problem").value.trim();

  if (!name) {
    showWarning(
      document.getElementById("name"),
      "Пожалуйста, введите ФИ преподавателя",
      "top",
      "50%",
      4000
    );
    return;
  }

  if (!place) {
    showWarning(
      document.getElementById("place"),
      "Пожалуйста, заполните поле",
      "top",
      "50%",
      4000
    );
    return;
  }

  if (!groupNumber) {
    showWarning(
      document.getElementById("group-number"),
      "Пожалуйста, заполните поле",
      "top",
      "50%",
      4000
    );
    return;
  }

  if (departament === "Отдел") {
    showWarning(
      document.querySelector("#teg"),
      "Пожалуйста, выберите отдел",
      "top",
      "50%",
      4000
    );
    return;
  }

  if (readyProblem === "Проблема") {
    showWarning(
      document.getElementById("problem"),
      "Пожалуйста, выберите готовую проблему или опишите свою",
      "top",
      "50%",
      4000
    );
    return;
  }

  if (!customProblem || customProblem.length < 5) {
    showWarning(
      document.getElementById("custom-problem"),
      "Пожалуйста, заполните поле",
      "top",
      "50%",
      4000
    );
    return;
  }

  let data = {
    initData: window.Telegram.WebApp.initData,
    "is-emergency": isEmergency,
    name: name,
    place: place,
    "group-number": groupNumber,
    departament: departament,
    "ready-problem": readyProblem !== "Проблема" ? readyProblem : "",
    "custom-problem": customProblem,
  };

  let button = document.getElementById("send");
  button.disabled = true;
  button.innerHTML = "";
  document.documentElement.style.setProperty("--main-spin-width", "5vw");
  document.documentElement.style.setProperty("--main-spin-heigt", "5vw");
  button.classList.add("spin-main");

  sendPostRequestAsync("/send-data", JSON.stringify(data))
    .then((result) => {
      button.classList.remove("spin-main");
      button.disabled = false;
      button.innerHTML = "Отправить";

      if (result.success) {
        showAlert(
          false,
          "Успех",
          "Заявка успешно отправлена",
          6
        );
        resetForm();
      } else {
        showAlert(
          true,
          "Ошибка",
          "Ошибка при отправке данных: " + result.response,
          5
        );
      }
    })
    .catch((error) => {
      button.classList.remove("spin-main");
      button.disabled = false;
      button.innerHTML = "Отправить";
      showAlert(false, "Ошибка", "Произошла ошибка: " + error.message, 5);
    });
}

function resetForm() {
  document.querySelector("#is-emergency").checked = false;
  document.querySelector("#name").value = "";
  document.querySelector("#place").value = "";
  document.querySelector("#group-number").value = "";
  document.querySelector("#teg").textContent = "Отдел";
  document
    .querySelectorAll(".problem .custom-select .select-selected")
    .forEach((el) => (el.textContent = "Проблема"));
  document.querySelector("#custom-problem").value = "";
}

// Функция для настройки кнопки отправки
function setMainButton() {
  let button = document.getElementById("send");
  button.addEventListener("click", handleSubmit);
}

// Функция для настройки кнопки отправки
function setMainButton() {
  let button = document.getElementById("send");
  button.addEventListener("click", handleSubmit);
}

// Вызываем setMainButton при загрузке DOM
document.addEventListener("DOMContentLoaded", setMainButton);

function organizeCustomSelects() {
  const organizedSelects = {};
  const customSelects = document.querySelectorAll(".custom-select");

  customSelects.forEach((select) => {
    const listItems = select.querySelectorAll("li");
    listItems.forEach((item) => {
      const tegIndex = item.getAttribute("data-teg-index");
      if (tegIndex !== null) {
        if (!organizedSelects[tegIndex]) {
          organizedSelects[tegIndex] = [];
        }
        if (!organizedSelects[tegIndex].includes(select)) {
          organizedSelects[tegIndex].push(select);
        }
      }
    });
  });

  return organizedSelects;
}

function getCustomSelectsByIndex(index) {
  return organizedCustomSelects[index] || [];
}

function extractTegNumber(tegElement) {
  if (!(tegElement instanceof HTMLElement)) {
    console.error("tegElement не является HTML элементом:", tegElement);
    return null;
  }
  const classList = tegElement.classList;
  for (let i = 0; i < classList.length; i++) {
    const match = classList[i].match(/^teg-(\d+)$/);
    if (match) {
      return parseInt(match[1], 10);
    }
  }
  console.error("Не найден класс teg-X в элементе:", tegElement);
  return null;
}

function toggleCustomSelectDisplay(selectedIndex) {
  Object.keys(organizedCustomSelects).forEach((index) => {
    const selects = organizedCustomSelects[index];
    selects.forEach((select) => {
      if (index === selectedIndex.toString()) {
        select.style.display = "flex";
      } else {
        select.style.display = "none";
      }
    });
  });
}

function setupSwipeHandler(handler) {
  let startX, startY, startTime;
  const minDistance = 100; // минимальное расстояние для свайпа (в пикселях)
  const maxTime = 300; // максимальное время для свайпа (в миллисекундах)
  const minSpeed = 0.5; // минимальная скорость свайпа (пиксели в миллисекунду)

  document.addEventListener(
    "touchstart",
    (e) => {
      const touch = e.touches[0];
      startX = touch.clientX;
      startY = touch.clientY;
      startTime = new Date().getTime();
    },
    false
  );

  document.addEventListener(
    "touchend",
    (e) => {
      if (!startX || !startY) return;

      const touch = e.changedTouches[0];
      const endX = touch.clientX;
      const endY = touch.clientY;
      const endTime = new Date().getTime();

      const distanceX = startX - endX;
      const distanceY = Math.abs(startY - endY);
      const time = endTime - startTime;
      const speed = distanceX / time;

      if (
        distanceX > minDistance &&
        distanceY < distanceX / 2 &&
        time < maxTime &&
        speed > minSpeed
      ) {
        handler({
          direction: "left",
          distance: distanceX,
          speed: speed,
          time: time,
        });
      }

      startX = startY = null;
    },
    false
  );
}

function sendGetRequestAsync(url, urlQuery) {
  return new Promise((resolve, reject) => {
    var Request = false;

    if (window.XMLHttpRequest) {
      Request = new XMLHttpRequest();
    } else if (window.ActiveXObject) {
      try {
        Request = new ActiveXObject("Microsoft.XMLHTTP");
      } catch (CatchException) {
        try {
          Request = new ActiveXObject("Msxml2.XMLHTTP");
        } catch (CatchException2) {
          Request = false;
        }
      }
    }

    if (!Request) {
      reject(new Error("Невозможно создать XMLHttpRequest"));
      return;
    }

    const fullUrl = urlQuery ? `${url}?${urlQuery}` : url;
    console.log("fullUrl    ", fullUrl);

    Request.open("GET", fullUrl, true);

    Request.onreadystatechange = function () {
      if (Request.readyState === 4) {
        if (Request.status === 200) {
          resolve({ success: true, response: Request.responseText });
        } else {
          let response;
          switch (Request.status) {
            case 404:
              response = "404 (Not Found)";
              break;
            case 403:
              response = "403 (Forbidden)";
              break;
            case 500:
              response = "500 (Internal Server Error)";
              break;
            default:
              response = `${Request.status} (${
                Request.statusText || "Unknown Error"
              })`;
          }
          resolve({ success: false, response: response });
        }
      }
    };

    Request.onerror = function () {
      reject(new Error("Network Error"));
    };

    try {
      Request.send();
    } catch (error) {
      reject(error);
    }
  });
}

function sendPostRequestAsync(url, data) {
  return new Promise((resolve, reject) => {
    var Request = false;

    if (window.XMLHttpRequest) {
      Request = new XMLHttpRequest();
    } else if (window.ActiveXObject) {
      try {
        Request = new ActiveXObject("Microsoft.XMLHTTP");
      } catch (CatchException) {
        try {
          Request = new ActiveXObject("Msxml2.XMLHTTP");
        } catch (CatchException2) {
          Request = false;
        }
      }
    }

    if (!Request) {
      reject(new Error("Невозможно создать XMLHttpRequest"));
      return;
    }

    console.log("URL:", url);
    console.log("Data:", data);

    Request.open("POST", url, true);

    Request.setRequestHeader(
      "Content-Type",
      "application/x-www-form-urlencoded"
    );

    Request.onreadystatechange = function () {
      if (Request.readyState === 4) {
        if (Request.status === 200) {
          resolve({ success: true, response: Request.responseText });
        } else {
          let response;
          switch (Request.status) {
            case 404:
              response = "404 (Not Found)";
              break;
            case 403:
              response = "403 (Forbidden)";
              break;
            case 500:
              response = "500 (Internal Server Error)";
              break;
            default:
              response = `${Request.status} (${
                Request.statusText || "Unknown Error"
              })`;
          }
          resolve({ success: false, response: response });
        }
      }
    };

    Request.onerror = function () {
      reject(new Error("Network Error"));
    };

    try {
      Request.send(data);
    } catch (error) {
      reject(error);
    }
  });
}

// Функция инициализации Telegram Web App
function initializeTg(maxAttempts, interval) {
  let attempts = 0;

  function tryInitialize() {
    if (window.Telegram && window.Telegram.WebApp) {
      const tg = window.Telegram.WebApp;
      const initData = tg.initData || tg.initDataUnsafe;
      if (initData) {
        console.log("initData получен:", initData);
        return { tg, initData };
      } else {
        console.warn("initData не доступен, пробуем снова...");
      }
    }
    return null;
  }

  return new Promise((resolve, reject) => {
    function attempt() {
      const result = tryInitialize();
      if (result) {
        resolve(result);
      } else if (attempts < maxAttempts) {
        attempts++;
        console.log(`Попытка ${attempts} из ${maxAttempts}`);
        setTimeout(attempt, interval);
      } else {
        reject(
          new Error(
            "Не удалось инициализировать Telegram Web App после " +
              maxAttempts +
              " попыток"
          )
        );
      }
    }
    attempt();
  });
}

// Функция для показа alert
function showAlert(message, duration, isError = false) {
  const alertContainer = document.getElementById("alert-container");
  const titleElement = document.getElementById("title");
  const descriptionElement = document.getElementById("error-description");

  titleElement.textContent = isError ? "ОШИБКА" : "УВЕДОМЛЕНИЕ";
  descriptionElement.textContent = message;

  alertContainer.classList.remove("off-screen");

  setTimeout(() => {
    alertContainer.classList.add("off-screen");
  }, duration * 1000);
}

// document.addEventListener("DOMContentLoaded", function () {
//   // Обработчик для кнопки закрытия alert
//   document
//     .getElementById("alert-close-button")
//     .addEventListener("click", function () {
//       document.getElementById("alert-container").classList.add("off-screen");
//     });
// });

// Alert анимация и логика
function animateLoading(durationInSeconds, callbackFunction) {
  const loadingElement = document.getElementById("loading");
  let startTime = null;

  function animate(currentTime) {
    if (!startTime) startTime = currentTime;
    const elapsedTime = currentTime - startTime;
    const progress = Math.min(elapsedTime / (durationInSeconds * 1000), 1);

    loadingElement.style.width = `${progress * 90}%`;

    if (progress < 1) {
      requestAnimationFrame(animate);
    } else {
      setTimeout(() => {
        // loadingElement.style.width = "0%";
        if (typeof callbackFunction === "function") {
          callbackFunction();
        }
      }, 100);
    }
  }

  requestAnimationFrame(animate);
}

function showAlert(isError, title, description, durationInSeconds) {
  const dialog = document.getElementById("dialog-alert");
  const titleElement = document.getElementById("alert-title");
  const descriptionElement = document.getElementById("alert-description");
  const loadingElement = document.getElementById("loading");

  if (isError) {
    titleElement.style.color = "rgb(255, 53, 53)";
  } else {
    titleElement.style.color = "rgb(62, 255, 62)";
  }

  titleElement.textContent = title;
  descriptionElement.textContent = description;

  // Возвращаем диалог на экран
  dialog.style.top = "50%";
  dialog.style.transform = "translateY(-50%)";

  dialog.showModal();
  dialog.classList.remove("hidden");

  animateLoading(durationInSeconds, () => {
    closeAlert();
  });
}

function closeAlert() {
  const dialog = document.getElementById("dialog-alert");
  dialog.classList.add("hidden");
  setTimeout(() => {
    dialog.close();
    // document.getElementById("loading").style.width = "0%";
    // Перемещаем диалог за пределы экрана
    dialog.style.top = "-100%";
  }, 300);
}

// document.addEventListener("DOMContentLoaded", function () {

// });
document.addEventListener("DOMContentLoaded", function () {
  const dialog = document.getElementById("dialog-alert");
  dialog.classList.add("hidden");
  dialog.style.top = "-100%";

  const checkbox = document.getElementById("is-emergency");
  if (checkbox) {
    checkbox.addEventListener("change", function () {
      console.log("Чекбокс изменен. Новое состояние:", this.checked);
    });
  } else {
    console.error("Чекбокс не найден");
  }
});

// Анимация диалогов
document.addEventListener("DOMContentLoaded", function () {
  const dialogs = document.querySelectorAll(".dialog");
  const openButtons = document.querySelectorAll(".button-hint");
  const closeButtons = document.querySelectorAll(".close-modal");

  function closeAllDialogs() {
    dialogs.forEach((dialog) => {
      if (dialog.open) {
        closeDialogWithAnimation(dialog);
      }
    });
  }

  function closeDialogWithAnimation(dialog) {
    dialog.classList.remove("animate-in");
    dialog.classList.add("animate-out");
    dialog.addEventListener("animationend", function handler() {
      dialog.close();
      dialog.classList.remove("animate-out");
      dialog.removeEventListener("animationend", handler);
    });
  }

  openButtons.forEach((button) => {
    button.addEventListener("click", (e) => {
      e.preventDefault();
      closeAllDialogs();
      const dialogId = button.getAttribute("data-dialog-id");
      const dialog = document.getElementById(dialogId);
      if (dialog) {
        dialog.showModal();
        dialog.classList.add("animate-in");
      }
    });
  });

  closeButtons.forEach((button) => {
    button.addEventListener("click", (e) => {
      e.preventDefault();
      const dialog = button.closest(".dialog");
      if (dialog) {
        closeDialogWithAnimation(dialog);
      }
    });
  });

  dialogs.forEach((dialog) => {
    dialog.addEventListener("close", () => {
      dialog.classList.remove("animate-in", "animate-out");
    });
  });

  window.addEventListener("click", (e) => {
    dialogs.forEach((dialog) => {
      if (
        dialog.open &&
        !dialog.contains(e.target) &&
        !e.target.closest(".button-hint")
      ) {
        closeDialogWithAnimation(dialog);
      }
    });
  });
});

// Свайп при фокусе
function scrollInputToMiddle() {
  const inputs = document.querySelectorAll("input, textarea");

  inputs.forEach((input) => {
    input.addEventListener("focus", function () {
      // Увеличиваем задержку до 1 секунды, чтобы дать время клавиатуре появиться
      setTimeout(() => {
        const viewportHeight = window.innerHeight;
        const inputRect = this.getBoundingClientRect();
        const inputTop = inputRect.top;
        const inputHeight = inputRect.height;

        // Вычисляем желаемую позицию прокрутки
        let targetScrollTop =
          window.pageYOffset + inputTop - viewportHeight / 2 + inputHeight / 2;

        // Проверяем, не пытаемся ли мы прокрутить дальше, чем возможно
        const maxScrollTop =
          document.documentElement.scrollHeight - viewportHeight;
        targetScrollTop = Math.min(targetScrollTop, maxScrollTop);

        // Выполняем прокрутку
        window.scrollTo({
          top: targetScrollTop,
          behavior: "smooth",
        });

        // Дополнительная проверка и корректировка через короткое время
        setTimeout(() => {
          const newInputRect = this.getBoundingClientRect();
          if (newInputRect.bottom > viewportHeight) {
            window.scrollBy({
              top: newInputRect.bottom - viewportHeight + 20, // 20px дополнительно для небольшого отступа
              behavior: "smooth",
            });
          }
        }, 300);
      }, 500); // Увеличенная задержка в 1 секунду
    });
  });
}

// Вызовите эту функцию после загрузки DOM
document.addEventListener("DOMContentLoaded", scrollInputToMiddle);

// Слушатель blur и focus для всех кнопок
document.addEventListener("DOMContentLoaded", function () {
  // Получаем все элементы input и textarea на странице
  const inputs = document.querySelectorAll("input, textarea");
  const emptySpace = document.getElementById("empty-space");

  // Добавляем обработчики событий blur и focus для каждого элемента
  inputs.forEach(function (input) {
    // Изначально скрываем пустой блок
    emptySpace.style.display = "none";

    // Обработчик события focus
    input.addEventListener("focus", function () {
      emptySpace.style.display = "block";
    });

    // Обработчик события blur
    input.addEventListener("blur", function () {
      emptySpace.style.display = "none";
    });

    // let button = document.getElementById("input-close-button");
    // input.addEventListener("blur", function () {
    //   button.style.display = "none";
    // });
  });
});
