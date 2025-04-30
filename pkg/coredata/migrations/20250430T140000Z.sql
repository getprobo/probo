ALTER TABLE mesures RENAME TO measures;
ALTER TABLE controls_mesures RENAME TO controls_measures;
ALTER TABLE risks_mesures RENAME TO risks_measures;
ALTER TABLE tasks RENAME COLUMN mesure_id TO measure_id;
ALTER TABLE controls_measures RENAME COLUMN mesure_id TO measure_id;
ALTER TABLE risks_measures RENAME COLUMN mesure_id TO measure_id;