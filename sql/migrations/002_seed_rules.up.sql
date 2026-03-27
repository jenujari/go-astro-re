INSERT INTO rules (id, name, category, status, tags, current_version, priority, is_modifier)
VALUES
    ('moon_in_rohini', 'Moon In Rohini', 'lunar_strength', 'active', '["moon","nakshatra","starter"]'::jsonb, '1.0.0', 10, FALSE),
    ('saturn_aspect_on_moon', 'Saturn Aspect On Moon', 'lunar_affliction', 'active', '["saturn","moon","aspect"]'::jsonb, '1.0.0', 20, FALSE),
    ('mars_in_own_sign', 'Mars In Own Sign', 'planetary_strength', 'active', '["mars","own-sign"]'::jsonb, '1.0.0', 30, FALSE),
    ('jupiter_exalted', 'Jupiter Exalted', 'planetary_strength', 'active', '["jupiter","exaltation"]'::jsonb, '1.0.0', 40, FALSE),
    ('sun_debilitated', 'Sun Debilitated', 'planetary_affliction', 'active', '["sun","debilitation"]'::jsonb, '1.0.0', 50, FALSE)
ON CONFLICT (id) DO NOTHING;

INSERT INTO rule_versions (rule_id, version, description, status, config, checksum)
VALUES
    ('moon_in_rohini', '1.0.0', 'Awards +3 when Moon is in Rohini', 'active', '{}'::jsonb, 'seed-moon-rohini-1'),
    ('saturn_aspect_on_moon', '1.0.0', 'Applies -4 when Saturn aspects Moon', 'active', '{}'::jsonb, 'seed-saturn-moon-1'),
    ('mars_in_own_sign', '1.0.0', 'Awards +2 when Mars is in own sign', 'active', '{}'::jsonb, 'seed-mars-own-sign-1'),
    ('jupiter_exalted', '1.0.0', 'Awards +5 when Jupiter is exalted', 'active', '{}'::jsonb, 'seed-jupiter-exalted-1'),
    ('sun_debilitated', '1.0.0', 'Applies -3 when Sun is debilitated', 'active', '{}'::jsonb, 'seed-sun-debilitated-1')
ON CONFLICT (rule_id, version) DO NOTHING;
