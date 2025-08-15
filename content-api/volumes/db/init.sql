CREATE TABLE IF NOT EXISTS content (
    id SERIAL PRIMARY KEY,
    slug TEXT NOT NULL,
    lang TEXT NOT NULL,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    UNIQUE(slug, lang)
    );
INSERT INTO content (slug, lang, title, body) VALUES
                                                  ('programs','ru','Образовательные программы',
                                                   'Список программ: Агрономия, Ветеринария, Инж.-техн., IT и др. Подробности: сайт/приёмка.'),
                                                  ('documents','ru','Документы для поступления',
                                                   'Паспорт/ID, Аттестат, Сертификат ЕНТ, Фото 3x4, Мед.справка 075-У, Заявление, и т.д.'),
                                                  ('grants','ru','Гранты',
                                                   'Гранты распределяются по результатам ЕНТ/квотам. Сроки подачи и проходные баллы — в приёмке.'),
                                                  ('dorm','ru','Общежитие',
                                                   'Места предоставляются в приоритетном порядке: иногородние/льготные категории. Узнайте очередность.'),
                                                  ('programs','kz','Білім беру бағдарламалары',
                                                   'Бағдарламалар тізімі: Агрономия, Ветеринария, Инж.-тех., IT және т.б. Толығырақ: сайт/қабылдау.'),
                                                  ('documents','kz','Құжаттар тізімі',
                                                   'Жеке куәлік, Аттестат, ҰБТ сертификаты, 3x4 фото, 075-У меданықтама, Өтініш және т.б.'),
                                                  ('grants','kz','Гранттар',
                                                   'Гранттар ҰБТ нәтижелері/квоталар бойынша бөлінеді. Өтініс мерзімдері мен өту балдары — қабылдауда.'),
                                                  ('dorm','kz','Жатақхана',
                                                   'Орындар басымдықпен беріледі: қаладан тыс/жеңілдік санаттары. Кезектілікті нақтылаңыз.')
ON CONFLICT DO NOTHING;
INSERT INTO content (slug, lang, title, body) VALUES
                                                  ('why-wkatu','ru','Почему WKATU','• Практико-ориентированное обучение\n• Сильные агро и инженерные направления\n• Стипендии и грантовые программы\n• Общежитие и студклубы\n• Партнёрства с работодателями'),
                                                  ('why-wkatu','kz','Неге WKATU','• Тәжірибеге бағытталған оқу\n• Күшті агро және инженерлік бағыттар\n• Стипендиялар мен гранттар\n• Жатақхана мен студенттік клубтар\n• Жұмыс берушілермен әріптестік'),

                                                  ('campus','ru','Студенческая жизнь в WKATU','Клубы и секции: IT, агротех, спорт, медиа. Регулярные мероприятия, волонтёрство, хакатоны. Спортзал и секции. Узнать актуальное — у студсовета.'),
                                                  ('campus','kz','WKATU студенттік өмірі','Клубтар мен секциялар: IT, агротех, спорт, медиа. Тұрақты іс-шаралар, волонтёрлік, хакатондар. Спортзал және секциялар. Актуалды — студенттер кеңесінде.');


CREATE TABLE "узлы_меню" (
    id            BIGSERIAL PRIMARY KEY,
    "code"        TEXT UNIQUE NOT NULL,
    "title_key"   TEXT NOT NULL,
    "slug"        TEXT,
    "parent_id"   BIGINT REFERENCES "узлы_меню"(id) ON DELETE SET NULL,
    "order"       INT NOT NULL DEFAULT 0,
    "active"      BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE "кнопки" (
    id BIGSERIAL PRIMARY KEY,
    node_id BIGINT NOT NULL REFERENCES узлы_меню(id) ON DELETE CASCADE,
    "text_key" TEXT NOT NULL,
    "next_node_id"  BIGINT REFERENCES "узлы_меню"(id) ON DELETE SET NULL,
    "order" INT NOT NULL DEFAULT 0
);

CREATE TABLE "переводы" (
    id      BIGSERIAL PRIMARY KEY,
    "lang"  TEXT NOT NULL,
    "key"   TEXT NOT NULL,
    "text"  TEXT NOT NULL,
    UNIQUE("lang","key")
);

CREATE INDEX ON "узлы_меню"("parent_id");
CREATE INDEX ON "кнопки"("node_id");
CREATE INDEX ON "переводы"("key","lang");

-- 02_seed.sql
INSERT INTO "узлы_меню"("code","title_key","order")
VALUES
    ('main','title.main',1),
    ('grants','title.grants',2),
    ('documents','title.documents',3),
    ('programs','title.programs',4),
    ('dorm','title.dorm',5);

UPDATE "узлы_меню" SET "slug"='grants'    WHERE "code"='grants';
UPDATE "узлы_меню" SET "slug"='documents' WHERE "code"='documents';
UPDATE "узлы_меню" SET "slug"='programs'  WHERE "code"='programs';
UPDATE "узлы_меню" SET "slug"='dorm'      WHERE "code"='dorm';

INSERT INTO "кнопки"("node_id","text_key","next_node_id","order")
SELECT m.id, s.key, n.id, s.ord
FROM (VALUES
          ('btn.programs','programs',1),
          ('btn.documents','documents',2),
          ('btn.grants','grants',3),
          ('btn.dorm','dorm',4)
     ) AS s(key,next,ord)
         JOIN "узлы_меню" m ON m."code"='main'
         JOIN "узлы_меню" n ON n."code"=s.next;

-- переводы (пример)
INSERT INTO "переводы"("lang","key","text") VALUES
                                                ('ru','title.main','Главное меню'),
                                                ('kz','title.main','Негізгі мәзір'),

                                                ('ru','title.programs','Образовательные программы'),
                                                ('kz','title.programs','Білім беру бағдарламалары'),

                                                ('ru','title.documents','Документы'),
                                                ('kz','title.documents','Құжаттар'),

                                                ('ru','title.grants','Гранты'),
                                                ('kz','title.grants','Гранттар'),

                                                ('ru','title.dorm','Общежитие'),
                                                ('kz','title.dorm','Жатақхана'),

                                                ('ru','btn.programs','🎓 Программы'),
                                                ('kz','btn.programs','🎓 Бағдарламалар'),

                                                ('ru','btn.documents','📑 Документы'),
                                                ('kz','btn.documents','📑 Құжаттар'),

                                                ('ru','btn.grants','🎁 Гранты'),
                                                ('kz','btn.grants','🎁 Гранттар'),

                                                ('ru','btn.dorm','🏠 Общежитие'),
                                                ('kz','btn.dorm','🏠 Жатақхана');
