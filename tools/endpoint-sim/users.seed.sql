-- ============================================================
-- Bot-bruger seed til endpoint-simulation
-- Password: Simulation2026!  (MD5 – opgraderes til bcrypt ved første login)
-- Kør med: psql $DATABASE_URL -f users.seed.sql
-- ============================================================

INSERT INTO users (username, email, password) VALUES
  ('bot_normal_01',  'bot_normal_01@sim.internal',  '1a98c5edf70365a9636d26b737163c47'),
  ('bot_normal_02',  'bot_normal_02@sim.internal',  '1a98c5edf70365a9636d26b737163c47'),
  ('bot_normal_03',  'bot_normal_03@sim.internal',  '1a98c5edf70365a9636d26b737163c47'),
  ('bot_normal_04',  'bot_normal_04@sim.internal',  '1a98c5edf70365a9636d26b737163c47'),
  ('bot_normal_05',  'bot_normal_05@sim.internal',  '1a98c5edf70365a9636d26b737163c47'),
  ('bot_normal_06',  'bot_normal_06@sim.internal',  '1a98c5edf70365a9636d26b737163c47'),
  ('bot_normal_07',  'bot_normal_07@sim.internal',  '1a98c5edf70365a9636d26b737163c47'),
  ('bot_normal_08',  'bot_normal_08@sim.internal',  '1a98c5edf70365a9636d26b737163c47'),
  ('bot_normal_09',  'bot_normal_09@sim.internal',  '1a98c5edf70365a9636d26b737163c47'),
  ('bot_normal_10',  'bot_normal_10@sim.internal',  '1a98c5edf70365a9636d26b737163c47'),
  ('bot_heavy_01',   'bot_heavy_01@sim.internal',   '1a98c5edf70365a9636d26b737163c47'),
  ('bot_heavy_02',   'bot_heavy_02@sim.internal',   '1a98c5edf70365a9636d26b737163c47'),
  ('bot_heavy_03',   'bot_heavy_03@sim.internal',   '1a98c5edf70365a9636d26b737163c47'),
  ('bot_heavy_04',   'bot_heavy_04@sim.internal',   '1a98c5edf70365a9636d26b737163c47'),
  ('bot_heavy_05',   'bot_heavy_05@sim.internal',   '1a98c5edf70365a9636d26b737163c47'),
  ('bot_session_01', 'bot_session_01@sim.internal', '1a98c5edf70365a9636d26b737163c47'),
  ('bot_session_02', 'bot_session_02@sim.internal', '1a98c5edf70365a9636d26b737163c47'),
  ('bot_session_03', 'bot_session_03@sim.internal', '1a98c5edf70365a9636d26b737163c47'),
  ('bot_session_04', 'bot_session_04@sim.internal', '1a98c5edf70365a9636d26b737163c47'),
  ('bot_session_05', 'bot_session_05@sim.internal', '1a98c5edf70365a9636d26b737163c47')
ON CONFLICT (username) DO NOTHING;
