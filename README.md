# Программный комплекс сбора метрик и алертинга

## Описание

Проект состоит из клиента и сервера. Клиент использует пакет `runtime` для получения параметров среды исполнения и посылает запросы с этими данными серверу, используя различные версии REST API (в разных учебных инкрементах серверу добавлялись дополнительные форматы API с сохранением возможности работы по старым форматам).  

Сервер хранит данные типа "ключ/значение", где ключ – название метрики. Поддерживаются различные типы хранилищ:  
– Хранилище в памяти (реализовано просто с помощью map). Дополнительно предусмотрена возможность сохранения и загрузки дампа этого хранилища на диск в виде json файла.  
– Хранилище в таблице базы данных Postgres

## Отработанные технологии:  

– Общие практики создания сервера на go, включая реализацию middlware, пакет chi.  
– Основы работы с СУБД Postgres с использованием универсального пакета sql.   
– Паттерны реализации многопоточности с использованием каналов и горутин, а также разрешение конфликтов доступа к разделяемым ресурсам.  
– Тестирование  
– Другие популярные библиотеки, включая сжатие, обработку json, логирование.