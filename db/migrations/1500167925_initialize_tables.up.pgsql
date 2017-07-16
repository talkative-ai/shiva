CREATE TABLE IF NOT EXISTS actors (
    id SERIAL PRIMARY KEY,
    title text,
    created_at timestamp DEFAULT current_timestamp
);
CREATE TABLE IF NOT EXISTS dialogs (
    id SERIAL PRIMARY KEY,
    title text,
    created_at timestamp DEFAULT current_timestamp
);
CREATE TABLE IF NOT EXISTS zones (
    id SERIAL PRIMARY KEY,
    title text,
    created_at timestamp DEFAULT current_timestamp
);
CREATE TABLE IF NOT EXISTS notes (
    id SERIAL PRIMARY KEY,
    title text,
    content text,
    created_at timestamp DEFAULT current_timestamp
);

CREATE TABLE IF NOT EXISTS projects (
    id SERIAL PRIMARY KEY,
    title text,
    owner_id text,
    start_zone_id integer references zones(id),
    created_at timestamp DEFAULT current_timestamp
);

CREATE TABLE IF NOT EXISTS project_actors (
    project_id integer NOT NULL references projects(id),
    actor_id integer NOT NULL references actors(id)
);
CREATE TABLE IF NOT EXISTS project_dialogs (
    project_id integer NOT NULL references projects(id),
    dialog_id integer NOT NULL references dialogs(id)
);
CREATE TABLE IF NOT EXISTS project_zones (
    project_id integer NOT NULL references projects(id),
    zone_id integer NOT NULL references zones(id)
);
CREATE TABLE IF NOT EXISTS project_notes (
    project_id integer NOT NULL references projects(id),
    note_id integer NOT NULL references notes(id)
);